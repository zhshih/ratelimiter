package distributed

import (
	"encoding/json"

	"github.com/hashicorp/raft"

	"github.com/zhshih/ratelimiter/internal/ratelimiter"
)

type RateLimiterSnapshot struct {
	data ratelimiter.RateLimitInfo
}

func (s *RateLimiterSnapshot) Persist(sink raft.SnapshotSink) error {
	data, err := json.Marshal(s.data)
	if err != nil {
		sink.Cancel()
		return err
	}

	if _, err := sink.Write(data); err != nil {
		sink.Cancel()
		return err
	}

	return sink.Close()
}

func (s *RateLimiterSnapshot) Release() {
}
