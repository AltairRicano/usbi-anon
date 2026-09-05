package httputil

import (
	"net/http/httptest"
	"testing"
)

// httputil no tenía pruebas (hallazgo de auditoría B9) a pesar de que ClientIP alimenta tanto al
// limitador de tasa como a la IP de evidencia legal del consentimiento de tutor. (Útil)

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

// TestClientIPIgnoresForwardedHeaders es una protección de regresión para la corrección B3:
// ClientIP nunca debe confiar en encabezados provistos por el cliente directamente, o un cliente directo
// podría falsificar su propia IP para la limitación de tasa y evidencia de auditoría de
// consentimiento de tutor. Confiar en encabezados (cuando se confirma que un proxy inverso los quita/pone)
// es trabajo exclusivo del middleware RealIP de chi, controlado por
// TRUST_PROXY_HEADERS — nunca de esta función. (Útil)
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
