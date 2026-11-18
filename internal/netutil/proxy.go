// Package netutil berisi parsing alamat proxy dan IP client.
package netutil

import (
	"log"
	"net"
	"strings"
)

// ParseTrustedProxies menerima daftar IP atau CIDR dipisah koma.
// X-Forwarded-For hanya dipercaya bila peer request berasal dari salah
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

// RequestIP mengambil IP dari "host:port", "[ipv6]:port", atau IP telanjang.
func RequestIP(addr string) net.IP {
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
