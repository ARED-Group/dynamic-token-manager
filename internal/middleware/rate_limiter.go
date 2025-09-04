package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter represents a simple rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	window   time.Duration
	cleanup  time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(window, cleanup time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		window:   window,
		cleanup:  cleanup,
	}
	
	// Start cleanup goroutine
	go rl.cleanupExpired()
	
	return rl
}

// RateLimitMiddleware returns a middleware that limits requests per IP
func (rl *RateLimiter) RateLimitMiddleware(limit int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			
			if !rl.Allow(ip, limit) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// Allow checks if a request is allowed for the given IP
func (rl *RateLimiter) Allow(ip string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	
	// Get existing requests for this IP
	requests := rl.requests[ip]
	
	// Filter out expired requests
	var validRequests []time.Time
	for _, req := range requests {
		if now.Sub(req) < rl.window {
			validRequests = append(validRequests, req)
		}
	}
	
	// Check if we're under the limit
	if len(validRequests) >= limit {
		return false
	}
	
	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[ip] = validRequests
	
	return true
}

// cleanupExpired removes expired entries from the rate limiter
func (rl *RateLimiter) cleanupExpired() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		
		for ip, requests := range rl.requests {
			var validRequests []time.Time
			for _, req := range requests {
				if now.Sub(req) < rl.window {
					validRequests = append(validRequests, req)
				}
			}
			
			if len(validRequests) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = validRequests
			}
		}
		
		rl.mu.Unlock()
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}