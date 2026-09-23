---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Proporciona los servicios y manejadores HTTP para la administración del banco de preguntas de registro y su configuración. Garantiza la invariante de mantener al menos 4 preguntas activas al desactivar o eliminar registros, restringe las operaciones exclusivamente al rol de administrador y audita de forma transaccional todos los cambios realizados con `logAudit` (ver [[backend/internal/quiz/bank.go.md|bank.go]]).

## Funciones

### AdminService.ListQuestions
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Obtiene la lista completa de preguntas de registro almacenadas en la base de datos y las transforma en su DTO de respuesta.

### AdminService.CreateQuestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida que el texto de la pregunta no esté vacío ni supere los 280 caracteres, inserta el registro con [[backend/internal/repository/quiz_queries.go.md#Queries.CreateRegistrationQuestion|Queries.CreateRegistrationQuestion]] y genera la entrada de auditoría dentro de una transacción con aislamiento serializable.

### AdminService.UpdateQuestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Ejecuta el reemplazo completo de la pregunta. En caso de desactivar una pregunta activa, verifica mediante bloqueo FOR UPDATE en transacción serializable que no se viole la regla del mínimo de 4 preguntas activas y registra el evento en auditoría.

### AdminService.DeleteQuestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Elimina una pregunta comprobando dentro de una transacción serializable que, si la pregunta estaba activa, se mantengan al menos 4 preguntas activas en el sistema, registrando la baja en auditoría.

### AdminService.GetSettings
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Consulta la configuración actual sobre el límite máximo de preguntas que se presentan a un usuario durante su registro.

### AdminService.UpdateSettings
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida que el límite máximo de preguntas a mostrar se encuentre en el rango de 4 a 10 y actualiza la configuración registrando en auditoría el estado anterior y el nuevo.

### Handler.ListQuestions
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Punto de entrada HTTP que verifica el rol de administrador en los claims JWT antes de delegar la consulta de preguntas al servicio.

### Handler.CreateQuestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Punto de entrada HTTP que autoriza únicamente administradores, decodifica estrictamente la solicitud JSON y retorna la pregunta recién creada.

### Handler.UpdateQuestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Punto de entrada HTTP que valida el rol de administrador y el UUID recibido en la URL antes de decodificar los datos y procesar la actualización.

### Handler.DeleteQuestion
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Punto de entrada HTTP que restringe el acceso a administradores, parsea el UUID de la URL y ejecuta el borrado devolviendo un estado sin contenido.

### Handler.GetSettings
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Punto de entrada HTTP que comprueba permisos de administrador y retorna la configuración del cuestionario de registro.

### Handler.UpdateSettings
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Punto de entrada HTTP que asegura rol de administrador, valida la estructura del payload JSON y delega la actualización de los ajustes.
