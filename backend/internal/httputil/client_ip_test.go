package httputil

import (
	"net/http/httptest"
	"testing"
)

// ClientIP alimenta tanto al limitador de tasa como a la IP de evidencia legal del consentimiento de tutor.
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
	req.RemoteAddr = "203.0.113.7" // malformado para SplitHostPort (sin puerto)

	if got := ClientIP(req); got != "203.0.113.7" {
		t.Fatalf("ClientIP() = %q, want %q", got, "203.0.113.7")
	}
}

// TestClientIPIgnoresForwardedHeaders verifica que ClientIP nunca confíe en encabezados
// provistos directamente por el cliente, evitando que un cliente directo pueda falsificar su propia IP
// para la limitación de tasa y evidencia de auditoría. Confiar en encabezados (cuando se confirma
// que un proxy inverso los gestiona) es responsabilidad exclusiva del middleware RealIP aguas arriba,
// controlado por TRUST_PROXY_HEADERS — nunca de esta función.
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
