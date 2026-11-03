package main

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter sederhana per IP: maksimal `limit` request per `window`.
// Cukup untuk deployment kecil; untuk multi-instance pakai Redis.
type RateLimiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	hits   map[string][]time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{limit: limit, window: window, hits: map[string][]time.Time{}}
}

func clientIP(r *http.Request) string {
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}
	return r.RemoteAddr
}

func (l *RateLimiter) Allow(ip string) bool {
	now := time.Now()
	cutoff := now.Add(-l.window)
	l.mu.Lock()
	defer l.mu.Unlock()
	kept := l.hits[ip][:0]
	for _, t := range l.hits[ip] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.limit {
		l.hits[ip] = kept
		return false
	}
	l.hits[ip] = append(kept, now)
	return true
}

func rateLimit(l *RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(clientIP(r)) {
			writeError(w, http.StatusTooManyRequests, "rate_limited", "terlalu banyak request, coba lagi sebentar lagi")
			return
		}
		next.ServeHTTP(w, r)
	})
}
