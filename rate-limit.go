package gonanoweb

import (
	"errors"
	"strings"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
	mu       sync.Mutex
	cleanup  *time.Ticker
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
		cleanup:  time.NewTicker(window),
	}

	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Filter old requests
	valid := rl.requests[ip][:0]
	for _, t := range rl.requests[ip] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	rl.requests[ip] = valid

	if len(valid) >= rl.limit {
		return false
	}

	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

func (rl *RateLimiter) cleanupLoop() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.window)
		for ip, times := range rl.requests {
			valid := times[:0]
			for _, t := range times {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = valid
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Stop() {
	rl.cleanup.Stop()
}

func RateLimitMiddleware(limiter *RateLimiter) Handler {
	return func(res *Response, req *Request) error {
		ip := strings.Split(req.conn.RemoteAddr().String(), ":")[0]
		if !limiter.Allow(ip) {
			res.ApiError(429, "Too Many Requests")
			return errors.New("rate limit exceeded")
		}
		return nil
	}
}

// Example:
// limiter := NewRateLimiter(100, time.Minute)
// defer limiter.Stop()
// server.UseMiddleware(RateLimitMiddleware(limiter))

