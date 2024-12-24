package distributed

import (
	"encoding/json"
	"io"
	"log"
	"sync"

	"github.com/hashicorp/raft"

	"github.com/zhshih/ratelimiter/internal/ratelimiter"
)

type ActionType uint8

const (
	Check ActionType = iota
	Increment
	Reset
)

type RateLimitCommand struct {
	Action    ActionType `json:"action"`
	ClientID  string     `json:"client_id"`
	ResetTime int64      `json:"reset_time"`
}

type ApplyResponse struct {
	Error error
	Data  interface{}
}

type RateLimiterFSM struct {
	mu          sync.Mutex
	rateLimiter *ratelimiter.RateLimiter
}

func NewRateLimiterFSM(limiter *ratelimiter.RateLimiter) *RateLimiterFSM {
	return &RateLimiterFSM{
		rateLimiter: limiter,
	}
}

func (fsm *RateLimiterFSM) Apply(raftLog *raft.Log) interface{} {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	var cmd RateLimitCommand
	if err := json.Unmarshal(raftLog.Data, &cmd); err != nil {
		return err
	}

	switch cmd.Action {
	case Check:
		remaining := fsm.rateLimiter.CheckQuota(cmd.ClientID)
		return &ApplyResponse{Error: nil, Data: remaining}
	case Increment:
		allowed := fsm.rateLimiter.AllowRequest(cmd.ClientID)
		return &ApplyResponse{Error: nil, Data: allowed}
	case Reset:
		fsm.rateLimiter.ResetQuota(cmd.ClientID)
		return &ApplyResponse{Error: nil}
	}
	return nil
}

func (fsm *RateLimiterFSM) Snapshot() (raft.FSMSnapshot, error) {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	rateLimitInfo := fsm.rateLimiter.GetRateLimitInfo()
	return &RateLimiterSnapshot{data: *rateLimitInfo}, nil
}

func (fsm *RateLimiterFSM) Restore(rc io.ReadCloser) error {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	defer func() {
		if err := rc.Close(); err != nil {
			log.Print("Close logstore for snapshot failed:", err)
		}
	}()

	log.Println("Read all message from snapshot")
	var totalRestored int

	decoder := json.NewDecoder(rc)
	for decoder.More() {
		payload := &RateLimitCommand{}
		err := decoder.Decode(payload)
		if err != nil {
			log.Print("Decode failed:", err)
			return err
		}

		ActionType := payload.Action
		log.Printf("ActionType = %d", ActionType)
		clientID := payload.ClientID
		log.Printf("clientID = %s", clientID)
		resetTime := payload.ResetTime
		log.Printf("resetTime = %d", resetTime)
		switch ActionType {
		case Increment:
			allowed := fsm.rateLimiter.AllowRequest(clientID)
			if !allowed {
				log.Printf("clientID %s is not allowed to request", clientID)
			}
		case Reset:
			fsm.rateLimiter.ResetQuota(clientID)
		}
		totalRestored++
	}

	_, err := decoder.Token()
	if err != nil {
		log.Print("Decode failed:", err)
		return err
	}

	log.Printf("Restore %d messages successfully in snapshot", totalRestored)
	return nil
}
