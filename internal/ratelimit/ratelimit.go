// Package ratelimit membatasi request per IP dengan jendela waktu sederhana.
// Cukup untuk satu instance; multi-instance membutuhkan penyimpanan bersama.
package ratelimit

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/enpidi26/vibelog/internal/netutil"
)

type Limiter struct {
	limit   int
	window  time.Duration
	trusted []*net.IPNet
	mu      sync.Mutex
	hits    map[string][]time.Time
	swept   time.Time
}

func New(limit int, window time.Duration, trusted []*net.IPNet) *Limiter {
	return &Limiter{
		limit:   limit,
		window:  window,
		trusted: trusted,
		hits:    map[string][]time.Time{},
		swept:   time.Now(),
	}
}

func (l *Limiter) isTrusted(ip net.IP) bool {
	for _, n := range l.trusted {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ClientIP mengembalikan IP klien sebenarnya. Header forwarded hanya dipakai
// bila peer adalah proxy tepercaya.
func (l *Limiter) ClientIP(r *http.Request) string {
	peer := netutil.RequestIP(r.RemoteAddr)
	if peer == nil {
		return r.RemoteAddr
	}
	if !l.isTrusted(peer) {
		return peer.String()
	}

	// XFF berisi rantai proxy dari kiri (klien) ke kanan (proxy terdekat).
	// Baca dari kanan: hop pertama yang tidak dipercaya adalah klien.
	if hops := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); hops != "" {
		parts := strings.Split(hops, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			ip := netutil.RequestIP(parts[i])
			if ip == nil {
				continue
			}
			if !l.isTrusted(ip) {
				return ip.String()
			}
			peer = ip
		}
		return peer.String()
	}

	if ip := netutil.RequestIP(r.Header.Get("X-Real-IP")); ip != nil {
		return ip.String()
	}
	return peer.String()
}

func (l *Limiter) Allow(ip string) bool {
	now := time.Now()
	cutoff := now.Add(-l.window)
	l.mu.Lock()
	defer l.mu.Unlock()

	// Buang entri basi secara berkala supaya map tidak tumbuh terus.
	if now.Sub(l.swept) > l.window {
		for key, times := range l.hits {
			if len(times) == 0 || !times[len(times)-1].After(cutoff) {
				delete(l.hits, key)
			}
		}
		l.swept = now
	}

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
