---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Suite de pruebas unitarias que valida la correcta extracción de la dirección IP del cliente a partir de `RemoteAddr`. Verifica el descarte del puerto en direcciones IPv4 e IPv6, el manejo de direcciones sin puerto y la invariante de seguridad que exige ignorar encabezados HTTP de reenvío (`X-Forwarded-For`, `X-Real-IP`, `True-Client-IP`) para prevenir la suplantación de IP en la limitación de tasa y en la evidencia legal de consentimiento.

Ejercita [[backend/internal/httputil/clientip.go.md#ClientIP|ClientIP]], la función que luego alimenta con la IP del cliente el sellado de evidencia de Service.CreateIncident.
