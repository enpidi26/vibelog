package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientIPTrustedProxies(t *testing.T) {
	trusted := ParseTrustedProxies("10.0.0.0/8, 192.0.2.7")
	limiter := NewRateLimiter(10, time.Minute, trusted)

	cases := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "untrusted peer ignores forwarded headers",
			remoteAddr: "203.0.113.9:5555",
			headers:    map[string]string{"X-Forwarded-For": "1.2.3.4", "X-Real-IP": "1.2.3.4"},
			want:       "203.0.113.9",
		},
		{
			name:       "trusted peer uses forwarded chain",
			remoteAddr: "10.1.2.3:5555",
			headers:    map[string]string{"X-Forwarded-For": "198.51.100.7, 10.0.0.4"},
			want:       "198.51.100.7",
		},
		{
			name:       "spoofed leftmost entry is not trusted",
			remoteAddr: "10.1.2.3:5555",
			headers:    map[string]string{"X-Forwarded-For": "1.1.1.1, 198.51.100.7"},
			want:       "198.51.100.7",
		},
		{
			name:       "trusted peer without forwarded headers",
			remoteAddr: "10.1.2.3:5555",
			want:       "10.1.2.3",
		},
		{
			name:       "single trusted ip entry",
			remoteAddr: "192.0.2.7:80",
			headers:    map[string]string{"X-Real-IP": "198.51.100.9"},
			want:       "198.51.100.9",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			req.RemoteAddr = tc.remoteAddr
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			if got := limiter.clientIP(req); got != tc.want {
				t.Fatalf("clientIP = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRateLimiterWindow(t *testing.T) {
	limiter := NewRateLimiter(2, 50*time.Millisecond, nil)
	if !limiter.Allow("198.51.100.7") || !limiter.Allow("198.51.100.7") {
		t.Fatal("first two requests should be allowed")
	}
	if limiter.Allow("198.51.100.7") {
		t.Fatal("third request inside the window should be rejected")
	}
	if !limiter.Allow("198.51.100.8") {
		t.Fatal("other clients must not be affected")
	}
	time.Sleep(60 * time.Millisecond)
	if !limiter.Allow("198.51.100.7") {
		t.Fatal("request after the window should be allowed again")
	}
}
