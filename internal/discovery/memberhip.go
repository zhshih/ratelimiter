package discovery

import (
	"fmt"
	"log"
	"net"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"

	"github.com/zhshih/ratelimiter/internal/config"
)

type Handler interface {
	Join(name, addr string) error
	Leave(name string) error
}

type DiscoveryAgent struct {
	Config  *config.ConfigMembership
	handler Handler
	serf    *serf.Serf
	events  chan serf.Event
}

func New(config *config.ConfigMembership, raftNode *raft.Raft) (*DiscoveryAgent, error) {
	m := &DiscoveryAgent{
		Config:  config,
		handler: newMemberHandler(raftNode),
	}
	if err := m.setupSerf(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *DiscoveryAgent) setupSerf() (err error) {
	addr, err := net.ResolveTCPAddr("tcp", m.Config.BindAddr)
	if err != nil {
		return err
	}
	config := serf.DefaultConfig()
	config.Init()
	config.MemberlistConfig.BindAddr = addr.IP.String()
	config.MemberlistConfig.BindPort = addr.Port
	m.events = make(chan serf.Event)
	config.EventCh = m.events
	config.Tags = m.Config.Tags
	config.NodeName = m.Config.NodeName
	m.serf, err = serf.Create(config)
	if err != nil {
		return err
	}

	go m.eventHandler()
	if m.Config.StartJoinAddrs != nil {
		for {
			_, err = m.serf.Join(m.Config.StartJoinAddrs, true)
			if err != nil {
				return err
			} else {
				break
			}
		}
	}
	return nil
}

func (m *DiscoveryAgent) eventHandler() {
	for e := range m.events {
		switch e.EventType() {
		case serf.EventMemberJoin:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					continue
				}
				m.handleJoin(member)
			}
		case serf.EventMemberLeave, serf.EventMemberFailed:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					return
				}
				m.handleLeave(member)
			}
		}
	}
}

func (m *DiscoveryAgent) handleJoin(member serf.Member) {
	log.Printf("member = %+v", member)
	if err := m.handler.Join(
		member.Name,
		member.Tags["raft_addr"],
	); err != nil {
		m.logError(err, "failed to join", member)
	}
}

func (m *DiscoveryAgent) handleLeave(member serf.Member) {
	if err := m.handler.Leave(
		member.Name,
	); err != nil {
		m.logError(err, "failed to leave", member)
	}
}

func (m *DiscoveryAgent) isLocal(member serf.Member) bool {
	return m.serf.LocalMember().Name == member.Name
}

func (m *DiscoveryAgent) logError(err error, msg string, member serf.Member) {
	if err == raft.ErrNotLeader {
		return
	}

	log.Printf(
		msg,
		fmt.Sprintf("name = %s", member.Name),
		fmt.Sprintf("raft_addr = %s", member.Tags["raft_addr"]),
	)
}

func newMemberHandler(raftNode *raft.Raft) Handler {
	return &memberHandler{
		raftNode: raftNode,
	}
}

type memberHandler struct {
	raftNode *raft.Raft
}

func (m *memberHandler) Join(id, addr string) error {
	configFuture := m.raftNode.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}
	serverID := raft.ServerID(id)
	serverAddr := raft.ServerAddress(addr)
	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == serverID || srv.Address == serverAddr {
			if srv.ID == serverID && srv.Address == serverAddr {
				return nil
			}
			log.Printf("remove server = %s", serverID)
			removeFuture := m.raftNode.RemoveServer(serverID, 0, 0)
			if err := removeFuture.Error(); err != nil {
				log.Printf("err = %s", err)
				return err
			}
		}
	}
	addFuture := m.raftNode.AddVoter(serverID, serverAddr, 0, 0)
	if err := addFuture.Error(); err != nil {
		log.Printf("err = %s", err)
		return err
	}
	return nil
}

func (m *memberHandler) Leave(id string) error {
	removeFuture := m.raftNode.RemoveServer(raft.ServerID(id), 0, 0)
	return removeFuture.Error()
}
