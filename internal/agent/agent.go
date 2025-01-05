package agent

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"github.com/zhshih/ratelimiter/internal/api"
	"github.com/zhshih/ratelimiter/internal/config"
	"github.com/zhshih/ratelimiter/internal/distributed"
	"github.com/zhshih/ratelimiter/internal/ratelimiter"
)

type Agent struct {
	cfgAPI      *config.ConfigAPI
	cfgRaft     *config.ConfigRaft
	ratelimiter *ratelimiter.RateLimiter
	raftNode    *raft.Raft
}

func NewAgent(cfgAPI *config.ConfigAPI, cfgRaft *config.ConfigRaft) *Agent {
	return &Agent{
		cfgAPI:  cfgAPI,
		cfgRaft: cfgRaft,
	}
}

func (a *Agent) Launch() {
	dataDir := a.cfgRaft.DataDir
	if _, err := os.Stat(dataDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(dataDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("Data directory %s already exists", dataDir)
	}
	limiter := ratelimiter.NewRateLimiter()
	a.ratelimiter = limiter

	if err := a.initRaft(dataDir); err != nil {
		log.Fatalf("Failed to create Raft node: %v", err)
	}

	a.launchAPI()
}

func (a *Agent) initRaft(dataDir string) error {
	raftNode, err := distributed.CreateRaftNode(
		&config.ConfigRaft{
			NodeID:   a.cfgRaft.NodeID,
			BindAddr: a.cfgRaft.BindAddr,
			DataDir:  dataDir,
		}, a.ratelimiter)
	if err != nil {
		return err
	}

	log.Printf("Raft node: %v created", raftNode)
	a.raftNode = raftNode
	return nil
}

func (a *Agent) launchAPI() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	raftHandler := &api.RaftHandler{
		RaftNode: a.raftNode,
	}
	router.POST("/raft/join", raftHandler.JoinRaftHandler)
	router.POST("/raft/remove", raftHandler.RemoveRaftHandler)
	router.GET("/raft/stats", raftHandler.StatsRaftHandler)

	apiHandler := &api.APIHandler{
		RateLimiter: a.ratelimiter,
		RaftNode:    a.raftNode,
	}
	router.GET("/rate/check", apiHandler.CheckQuotaHandler)
	router.POST("/rate/increment", apiHandler.IncrementQuotaHandler)
	router.POST("/rate/reset", apiHandler.ResetQuotaHandler)

	serverPort := fmt.Sprintf(":%d", a.cfgAPI.Port)
	log.Printf("Rate limiter running on %s", serverPort)
	if err := router.Run(serverPort); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
