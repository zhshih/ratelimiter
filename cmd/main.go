package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"

	"github.com/zhshih/ratelimiter/internal/agent"
	"github.com/zhshih/ratelimiter/internal/config"
)

type configRaft struct {
	NodeID    string `mapstructure:"node_id"`
	Port      int    `mapstructure:"port"`
	VolumeDir string `mapstructure:"volume_dir"`
}
type configServer struct {
	Port int `mapstructure:"port"`
}
type cfg struct {
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

	conf := cfg{
		Server: configServer{
			Port: v.GetInt(serverPort),
		},
		Raft: configRaft{
			NodeID:    v.GetString(raftNodeId),
			Port:      v.GetInt(raftPort),
			VolumeDir: v.GetString(raftVolDir),
		},
	}

	agent := agent.NewAgent(
		&config.ConfigAPI{
			Port: conf.Server.Port,
		}, &config.ConfigRaft{
			NodeID:   conf.Raft.NodeID,
			BindAddr: fmt.Sprintf("127.0.0.1:%d", conf.Raft.Port),
			DataDir:  conf.Raft.VolumeDir,
		},
	)
	agent.Launch()
}
