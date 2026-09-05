// Package httputil contiene pequeños helpers HTTP sin dependencias compartidos entre
// los paquetes de transporte y lógica de negocio. (Relleno)
package httputil

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP devuelve la dirección peer del socket para r, sin confiar en ningún
// encabezado provisto por el cliente. Los llamadores que necesiten confiar en los
// encabezados forwarded-for de un proxy inverso deben depender de que el middleware RealIP esté habilitado
// aguas arriba (el cual reescribe r.RemoteAddr) en lugar de volver a parsear encabezados aquí. (Útil)
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return strings.TrimSpace(host)
}
