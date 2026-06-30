package middleware

import (
	"net/http"
	"sync"
	"time"

	"app/internal/interfaces/http/response"
	"app/pkg/apperror"
	"golang.org/x/time/rate"
)

// RateLimiter holds the visitors map and a mutex for concurrent access.
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     r,
		burst:    b,
	}
	
	// Background cleanup of old visitors
	go rl.cleanupVisitors()
	
	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = &visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit is a middleware that applies rate limiting based on the remote IP address.
func RateLimit(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract IP from RemoteAddr or Headers
			ip := r.RemoteAddr
			if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				ip = forwardedFor
			} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				ip = realIP
			}
			
			limiter := rl.getLimiter(ip)
			if !limiter.Allow() {
				response.Error(w, r, nil, apperror.New(apperror.CodeRateLimited, "too many requests, please try again later", http.StatusTooManyRequests))
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}
