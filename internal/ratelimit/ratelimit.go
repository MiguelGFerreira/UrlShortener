// Package ratelimit provides a simple per-client token-bucket HTTP middleware.
package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// bucket tracks the available tokens for a single client.
type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// Limiter throttles requests per client IP using a token bucket that refills at
// `rate` tokens per second up to `burst` tokens.
type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64
	burst   float64
}

// New creates a Limiter that allows an initial burst of requests and then
// refills at ratePerSec tokens per second. A background goroutine evicts
// clients that have been idle for a while.
func New(ratePerSec, burst float64) *Limiter {
	l := &Limiter{
		buckets: make(map[string]*bucket),
		rate:    ratePerSec,
		burst:   burst,
	}
	go l.cleanupLoop()
	return l
}

// allow reports whether a request from key may proceed, consuming a token.
func (l *Limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &bucket{tokens: l.burst - 1, lastSeen: now}
		return true
	}

	// Refill proportionally to the time elapsed since the last request.
	b.tokens += now.Sub(b.lastSeen).Seconds() * l.rate
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.lastSeen = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Middleware wraps next, responding 429 when the client exceeds its rate.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(clientIP(r)) {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "Error: rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// cleanupLoop periodically discards buckets that have been idle for 10 minutes.
func (l *Limiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		l.mu.Lock()
		for key, b := range l.buckets {
			if time.Since(b.lastSeen) > 10*time.Minute {
				delete(l.buckets, key)
			}
		}
		l.mu.Unlock()
	}
}

// clientIP extracts the client's IP address from the request.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
