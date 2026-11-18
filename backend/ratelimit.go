package main

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter sederhana per IP: maksimal `limit` request per `window`.
// Cukup untuk deployment satu instance; untuk multi-instance pakai Redis.
type RateLimiter struct {
	limit   int
	window  time.Duration
	trusted []*net.IPNet
	mu      sync.Mutex
	hits    map[string][]time.Time
	swept   time.Time
}

// ParseTrustedProxies menerima daftar IP atau CIDR yang dipisah koma.
// X-Forwarded-For hanya dipercaya kalau peer request datang dari salah
// satunya, sehingga header palsu dari klien langsung tidak mempan.
func ParseTrustedProxies(raw string) []*net.IPNet {
	var nets []*net.IPNet
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "/") {
			_, n, err := net.ParseCIDR(part)
			if err != nil {
				log.Printf("TRUSTED_PROXIES: skipping invalid CIDR %q", part)
				continue
			}
			nets = append(nets, n)
			continue
		}
		ip := net.ParseIP(part)
		if ip == nil {
			log.Printf("TRUSTED_PROXIES: skipping invalid IP %q", part)
			continue
		}
		bits := 128
		if v4 := ip.To4(); v4 != nil {
			ip = v4
			bits = 32
		}
		nets = append(nets, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
	}
	return nets
}

func NewRateLimiter(limit int, window time.Duration, trusted []*net.IPNet) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  window,
		trusted: trusted,
		hits:    map[string][]time.Time{},
		swept:   time.Now(),
	}
}

func (l *RateLimiter) isTrusted(ip net.IP) bool {
	for _, n := range l.trusted {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func (l *RateLimiter) clientIP(r *http.Request) string {
	peer := requestIP(r.RemoteAddr)
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
			ip := requestIP(parts[i])
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

	if ip := requestIP(r.Header.Get("X-Real-IP")); ip != nil {
		return ip.String()
	}
	return peer.String()
}

func requestIP(addr string) net.IP {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil
	}
	if host, _, err := net.SplitHostPort(addr); err == nil {
		addr = host
	}
	addr = strings.Trim(addr, "[]")
	return net.ParseIP(addr)
}

func (l *RateLimiter) Allow(ip string) bool {
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

func rateLimit(l *RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(l.clientIP(r)) {
			writeError(w, http.StatusTooManyRequests, "rate_limited", "terlalu banyak request, coba lagi sebentar lagi")
			return
		}
		next.ServeHTTP(w, r)
	})
}
