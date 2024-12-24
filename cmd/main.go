package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/zhshih/ratelimiter/internal/api"
	"github.com/zhshih/ratelimiter/internal/distributed"
	"github.com/zhshih/ratelimiter/internal/ratelimiter"
)

type configRaft struct {
	NodeID    string `mapstructure:"node_id"`
	Port      int    `mapstructure:"port"`
	VolumeDir string `mapstructure:"volume_dir"`
}
type configServer struct {
	Port int `mapstructure:"port"`
}
type config struct {
	Server configServer `mapstructure:"server"`
	Raft   configRaft   `mapstructure:"raft"`
}

const (
	serverPort = "SERVER_PORT"
	raftNodeId = "RAFT_NODE_ID"
	raftPort   = "RAFT_PORT"
	raftVolDir = "RAFT_VOL_DIR"
)

var confKeys = []string{
	serverPort,
	raftNodeId,
	raftPort,
	raftVolDir,
}

func main() {
	var v = viper.New()
	v.AutomaticEnv()
	if err := v.BindEnv(confKeys...); err != nil {
		log.Fatal(err)
		return
	}

	conf := config{
		Server: configServer{
			Port: v.GetInt(serverPort),
		},
		Raft: configRaft{
			NodeID:    v.GetString(raftNodeId),
			Port:      v.GetInt(raftPort),
			VolumeDir: v.GetString(raftVolDir),
		},
	}

	bindAddr := fmt.Sprintf("127.0.0.1:%d", conf.Raft.Port)
	dataDir := conf.Raft.VolumeDir
	if _, err := os.Stat(dataDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(dataDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("Data directory %s already exists", dataDir)
	}
	rl := ratelimiter.NewRateLimiter()
	raftNode, err := distributed.CreateRaftNode(conf.Raft.NodeID, dataDir, bindAddr, rl)
	if err != nil {
		log.Fatalf("Failed to create Raft node: %v", err)
	}

	log.Printf("Raft node: %v created", raftNode)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	raftHandler := &api.RaftHandler{
		RaftNode: raftNode,
	}
	router.POST("/raft/join", raftHandler.JoinRaftHandler)
	router.POST("/raft/remove", raftHandler.RemoveRaftHandler)
	router.GET("/raft/stats", raftHandler.StatsRaftHandler)

	apiHandler := &api.APIHandler{
		RateLimiter: rl,
		RaftNode:    raftNode,
	}
	router.GET("/rate/check", apiHandler.CheckQuotaHandler)
	router.POST("/rate/increment", apiHandler.IncrementQuotaHandler)
	router.POST("/rate/reset", apiHandler.ResetQuotaHandler)

	serverPort := fmt.Sprintf(":%d", conf.Server.Port)
	log.Printf("Rate limiter running on %s", serverPort)
	if err := router.Run(serverPort); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
