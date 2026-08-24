package httputil

import (
	"net/http/httptest"
	"testing"
)

// httputil had zero tests (audit finding B9) despite ClientIP feeding both the
// rate limiter and the tutor-consent legal-evidence IP.

func TestClientIPSplitsHostPort(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.7:54321"

	if got := ClientIP(req); got != "203.0.113.7" {
		t.Fatalf("ClientIP() = %q, want %q", got, "203.0.113.7")
	}
}

func TestClientIPHandlesIPv6WithPort(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "[2001:db8::1]:443"

	if got := ClientIP(req); got != "2001:db8::1" {
		t.Fatalf("ClientIP() = %q, want %q", got, "2001:db8::1")
	}
}

func TestClientIPFallsBackWhenNoPort(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.7" // malformed for SplitHostPort (no port)

	if got := ClientIP(req); got != "203.0.113.7" {
		t.Fatalf("ClientIP() = %q, want %q", got, "203.0.113.7")
	}
}

// TestClientIPIgnoresForwardedHeaders is a regression guard for the B3 fix:
// ClientIP must never trust client-supplied headers directly, or a direct
// client could spoof its own IP for rate limiting and tutor-consent audit
// evidence. Header-trusting (when a reverse proxy is confirmed to strip/set
// them) is exclusively chi's RealIP middleware's job, gated by
// TRUST_PROXY_HEADERS — never this function.
func TestClientIPIgnoresForwardedHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.7:54321"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	req.Header.Set("X-Real-IP", "5.6.7.8")
	req.Header.Set("True-Client-IP", "9.9.9.9")

	if got := ClientIP(req); got != "203.0.113.7" {
		t.Fatalf("ClientIP() = %q, want the socket peer %q (headers must be ignored)", got, "203.0.113.7")
	}
}
