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
	"github.com/zhshih/ratelimiter/internal/discovery"
	"github.com/zhshih/ratelimiter/internal/distributed"
	"github.com/zhshih/ratelimiter/internal/ratelimiter"
)

type Agent struct {
	cfgAPI        *config.ConfigAPI
	cfgRaft       *config.ConfigRaft
	cfgMemberShip *config.ConfigMembership
	ratelimiter   *ratelimiter.RateLimiter
	raftNode      *raft.Raft
	membership    *discovery.DiscoveryAgent
}

func NewAgent(cfgAPI *config.ConfigAPI, cfgRaft *config.ConfigRaft, cfgMemberShip *config.ConfigMembership) *Agent {
	return &Agent{
		cfgAPI:        cfgAPI,
		cfgRaft:       cfgRaft,
		cfgMemberShip: cfgMemberShip,
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

	if err := a.initMembership(); err != nil {
		log.Fatalf("Failed to create membership: %v", err)
	}

	if err := a.launchAPI(); err != nil {
		log.Fatalf("Failed to launch API Server: %v", err)
	}
}

func (a *Agent) initRaft(dataDir string) error {
	raftNode, err := distributed.NewRaft(
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

func (a *Agent) initMembership() error {
	membership, err := discovery.New(a.cfgMemberShip, a.raftNode)
	if err != nil {
		return err
	}
	a.membership = membership
	return nil
}

func (a *Agent) launchAPI() error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	raftHandler := &api.RaftHandler{
		RaftNode: a.raftNode,
	}

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
		return err
	}
	return nil
}
