package ratelimiter

import (
	"sync"
	"time"
)

type RateLimitInfo struct {
	Info map[string]*clientQuota
}

type RateLimiter struct {
	limits *RateLimitInfo
	mu     sync.Mutex
}

type clientQuota struct {
	limit     int
	count     int
	resetTime time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits: &RateLimitInfo{
			Info: make(map[string]*clientQuota),
		},
	}
}

func (rl *RateLimiter) GetRateLimitInfo() *RateLimitInfo {
	return rl.limits
}

func (rl *RateLimiter) CheckQuota(clientID string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	quota, exists := rl.limits.Info[clientID]
	remaining := 0
	if exists {
		remaining = quota.limit - quota.count
		if remaining < 0 {
			remaining = 0
		}
	} else {
		remaining = 10
	}
	return remaining
}

func (rl *RateLimiter) AllowRequest(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	quota, exists := rl.limits.Info[clientID]
	if !exists {
		quota = &clientQuota{
			limit:     10,
			count:     0,
			resetTime: time.Now().Add(1 * time.Minute),
		}
		rl.limits.Info[clientID] = quota
	}

	if time.Now().After(quota.resetTime) {
		quota.count = 0
		quota.resetTime = time.Now().Add(1 * time.Minute)
	}

	if quota.count < quota.limit {
		quota.count++
		return true
	}

	return false
}

func (rl *RateLimiter) ResetQuota(clientID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if quota, exists := rl.limits.Info[clientID]; exists {
		quota.count = 0
		quota.resetTime = time.Now().Add(1 * time.Minute)
	}
}
