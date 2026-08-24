// Package httputil holds small, dependency-free HTTP helpers shared across
// transport and business-logic packages.
package httputil

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP returns the socket peer address for r, without trusting any
// client-supplied header. Callers that need to trust a reverse proxy's
// forwarded-for headers should rely on the RealIP middleware being enabled
// upstream (which rewrites r.RemoteAddr) rather than re-parsing headers here.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return strings.TrimSpace(host)
}
