package ratelimiter

import (
	"log"
	"sync"
	"time"
)

type RateLimitInfo struct {
	Info map[string]*ClientRateLimit
}

type RateLimiter struct {
	limits *RateLimitInfo
	mu     sync.Mutex
}

type ClientRateLimit struct {
	Quota       *clientQuota
	tokenBucket *TokenBucket
}

type clientQuota struct {
	limit     int
	count     int
	resetTime time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits: &RateLimitInfo{
			Info: make(map[string]*ClientRateLimit),
		},
	}
}

func (rl *RateLimiter) GetRateLimitInfo() *RateLimitInfo {
	return rl.limits
}

func (rl *RateLimiter) CheckQuota(clientID string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	clientRateLimit, exists := rl.limits.Info[clientID]
	remaining := 0
	if exists {
		resetTime := clientRateLimit.Quota.resetTime
		if time.Now().After(resetTime) {
			log.Printf("Ratelimit is reset to clientID = %s", clientID)
			clientRateLimit = rl.resetRateLimit()
			rl.limits.Info[clientID] = clientRateLimit
		}

		remaining = clientRateLimit.Quota.limit - clientRateLimit.Quota.count
		remaining = max(0, remaining)
	} else {
		remaining = 10
	}
	return remaining
}

func (rl *RateLimiter) AllowRequest(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	clientRateLimit, exists := rl.limits.Info[clientID]
	if !exists {
		clientRateLimit = rl.resetRateLimit()
		rl.limits.Info[clientID] = clientRateLimit
	}

	resetTime := clientRateLimit.Quota.resetTime
	if time.Now().After(resetTime) {
		log.Printf("Ratelimit is reset to clientID = %s", clientID)
		rl.limits.Info[clientID] = rl.resetRateLimit()
	}

	quota := clientRateLimit.Quota
	if quota.count < quota.limit && clientRateLimit.tokenBucket.tryConsume() {
		quota.count++
		return true
	}

	return false
}

func (rl *RateLimiter) ResetQuota(clientID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if _, exists := rl.limits.Info[clientID]; exists {
		log.Printf("Ratelimit is reset to clientID = %s", clientID)
		rl.limits.Info[clientID] = rl.resetRateLimit()
	}

}

func (rl *RateLimiter) resetRateLimit() *ClientRateLimit {
	return &ClientRateLimit{
		Quota: &clientQuota{
			limit:     10,
			count:     0,
			resetTime: time.Now().Add(1 * time.Minute),
		},
		tokenBucket: NewTokenBucket(10, 1),
	}
}
