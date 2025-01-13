package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"

	"github.com/zhshih/ratelimiter/internal/agent"
	"github.com/zhshih/ratelimiter/internal/config"
)

type configRaft struct {
	Port      int    `mapstructure:"port"`
	VolumeDir string `mapstructure:"volume_dir"`
}

type configServer struct {
	Port int `mapstructure:"port"`
}

type configDiscovery struct {
	Port int `mapstructure:"port"`
}

type cfg struct {
	NodeID            string          `mapstructure:"node_id"`
	Server            configServer    `mapstructure:"server"`
	Raft              configRaft      `mapstructure:"raft"`
	Discovery         configDiscovery `mapstructure:"discovery"`
	DiscoveryClusters []string        `mapstructure:"discoveryClusters"`
}

const (
	nodeId            = "NODE_ID"
	serverPort        = "SERVER_PORT"
	raftPort          = "RAFT_PORT"
	raftVolDir        = "RAFT_VOL_DIR"
	discoveryPort     = "DISCOVERY_PORT"
	discoveryClusters = "CLUSTERS"
)

var confKeys = []string{
	serverPort,
	nodeId,
	raftPort,
	raftVolDir,
	discoveryPort,
	discoveryClusters,
}

func main() {
	var v = viper.New()
	v.AutomaticEnv()
	if err := v.BindEnv(confKeys...); err != nil {
		log.Fatal(err)
		return
	}

	clusters := v.GetString(discoveryClusters)
	clusterList := strings.Split(clusters, ",")
	conf := cfg{
		NodeID: v.GetString(nodeId),
		Server: configServer{
			Port: v.GetInt(serverPort),
		},
		Raft: configRaft{
			Port:      v.GetInt(raftPort),
			VolumeDir: v.GetString(raftVolDir),
		},
		Discovery: configDiscovery{
			Port: v.GetInt(discoveryPort),
		},
		DiscoveryClusters: clusterList,
	}

	bindAddr := fmt.Sprintf("127.0.0.1:%d", conf.Raft.Port)
	agent := agent.NewAgent(
		&config.ConfigAPI{
			Port: conf.Server.Port,
		}, &config.ConfigRaft{
			NodeID:   conf.NodeID,
			BindAddr: bindAddr,
			DataDir:  conf.Raft.VolumeDir,
		}, &config.ConfigMembership{
			NodeName: conf.NodeID,
			BindAddr: fmt.Sprintf("127.0.0.1:%d", conf.Discovery.Port),
			Tags: map[string]string{
				"raft_addr": bindAddr,
			},
			StartJoinAddrs: conf.DiscoveryClusters,
		},
	)
	agent.Launch()
}
