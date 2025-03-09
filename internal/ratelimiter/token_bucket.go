package ratelimiter

import (
	"time"
)

type TokenBucket struct {
	tokens         int
	maxTokens      int
	refillRate     int
	lastRefillTime time.Time
}

func NewTokenBucket(maxTokens, refillRate int) *TokenBucket {
	return &TokenBucket{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()
	newTokens := int(elapsed) * tb.refillRate
	if newTokens > 0 {
		tb.tokens = tb.tokens + newTokens
		if tb.tokens > tb.maxTokens {
			tb.tokens = tb.maxTokens
		}
		tb.lastRefillTime = now
	}
}

func (tb *TokenBucket) tryConsume() bool {
	tb.refill()
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}
