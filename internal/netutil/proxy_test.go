package netutil

import "testing"

func TestParseTrustedProxies(t *testing.T) {
	nets := ParseTrustedProxies("10.0.0.0/8, 192.0.2.7, 2001:db8::/32, invalid, 10.0.0.0/99")
	if len(nets) != 3 {
		t.Fatalf("expected 3 valid entries, got %d: %v", len(nets), nets)
	}
}

func TestRequestIP(t *testing.T) {
	cases := map[string]string{
		"203.0.113.9:5555":  "203.0.113.9",
		"[2001:db8::1]:443": "2001:db8::1",
		"198.51.100.7":      "198.51.100.7",
		" 192.0.2.7 ":       "192.0.2.7",
		"":                  "",
		"not-an-ip":         "",
	}
	for in, want := range cases {
		got := RequestIP(in)
		if want == "" {
			if got != nil {
				t.Fatalf("RequestIP(%q) = %v, want nil", in, got)
			}
			continue
		}
		if got == nil || got.String() != want {
			t.Fatalf("RequestIP(%q) = %v, want %s", in, got, want)
		}
	}
}
