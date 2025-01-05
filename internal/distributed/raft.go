package distributed

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"

	"github.com/zhshih/ratelimiter/internal/config"
	"github.com/zhshih/ratelimiter/internal/ratelimiter"
)

const (
	// The maxPool controls how many connections we will pool.
	maxPool = 3

	// The timeout is used to apply I/O deadlines. For InstallSnapshot, we multiply
	// the timeout by (SnapshotSize / TimeoutScale).
	// https://github.com/hashicorp/raft/blob/v1.1.2/net_transport.go#L177-L181
	tcpTimeout = 10 * time.Second

	// The `retain` parameter controls how many
	// snapshots are retained. Must be at least 1.
	raftSnapShotRetain = 2

	// raftLogCacheSize is the maximum number of logs to cache in-memory.
	// This is used to reduce disk I/O for the recently committed entries.
	raftLogCacheSize = 512
)

type RaftNodeInfo struct {
	Addr string
	Node *raft.Raft
}

func CreateRaftNode(cfg *config.ConfigRaft, limiter *ratelimiter.RateLimiter) (*raft.Raft, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(cfg.NodeID)

	logStore, err := raftboltdb.NewBoltStore(cfg.DataDir + "/raft-log.bolt")
	if err != nil {
		return nil, err
	}

	cacheStore, err := raft.NewLogCache(raftLogCacheSize, logStore)
	if err != nil {
		return nil, err
	}

	snapshotStore, err := raft.NewFileSnapshotStore(cfg.DataDir, raftSnapShotRetain, os.Stderr)
	if err != nil {
		return nil, err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", cfg.BindAddr)
	if err != nil {
		log.Fatal(err)
	}

	transport, err := raft.NewTCPTransport(cfg.BindAddr, tcpAddr, maxPool, tcpTimeout, os.Stdout)
	if err != nil {
		return nil, err
	}

	fsm := NewRateLimiterFSM(limiter)

	raftNode, err := raft.NewRaft(config, fsm, cacheStore, logStore, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(cfg.NodeID),
				Address: transport.LocalAddr(),
			},
		},
	}

	raftNode.BootstrapCluster(configuration)

	return raftNode, nil
}
