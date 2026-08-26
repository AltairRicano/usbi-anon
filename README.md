# USBI-Anon

Variante de la plataforma educativa gamificada USBI (Universidad Veracruzana)
en la que cada persona usuaria queda identificada únicamente por un **UUID**
en el sistema principal. En vez de correo electrónico y contraseña, el
registro se resuelve con un breve cuestionario de gustos no sensibles (color
favorito, animal favorito, etc.) del que el sistema deriva automáticamente un
nickname y una contraseña — nadie escribe su nombre, correo ni teléfono en
ninguna parte del sistema.

El sistema conserva las mismas capacidades funcionales que la propuesta
original: catálogo de secciones y niveles, minijuegos, experiencia (XP),
insignias, rachas, progreso persistente entre dispositivos y panel de
administración. El cambio de fondo es de arquitectura de identidad y de
modelo de alojamiento, no de alcance funcional.

Licencia: Apache 2.0. Consulta [LICENSE](LICENSE).

## Estado actual

El backend (Go) implementa el esquema de base de datos unificado, el registro
en tres pasos, login/sesión, solicitudes ARCO y los paneles de administración
de cuentas y del banco de preguntas de registro. El frontend (React + TypeScript
+ Vite) cubre esos mismos flujos: registro, inicio de sesión y los dos paneles
de administración. El resto de la experiencia de juego (progreso, niveles,
minijuegos) todavía no se ha portado a este proyecto.

## Arquitectura de datos

A diferencia de un sistema tradicional, **una sola base de datos** contiene
tanto las credenciales de acceso como el progreso y contenido educativo,
indexados exclusivamente por UUID. Ninguna tabla contiene nombre, correo,
teléfono ni cualquier otro dato de contacto: el nickname que identifica a
cada cuenta se genera a partir de fragmentos de las respuestas al
cuestionario de registro, no de un dato personal.

Retirar contenido educativo entre temporadas (rotación de niveles) preserva
siempre la experiencia (XP) y los contadores de progreso ya ganados por cada
jugador, aunque el nivel original se elimine para liberar espacio.

## Identidad visual

El proyecto se presenta como producto de la Universidad Veracruzana / USBI y
sigue las convenciones institucionales de color, contraste y accesibilidad
documentadas en [`plan/Convenciones_de_color_UV.md`](plan/Convenciones_de_color_UV.md).

## Estructura del proyecto

```text
.
├── backend/                            # API en Go (chi router, PostgreSQL)
│   ├── cmd/
│   │   └── hash_password/              # binario standalone para el bootstrap del primer admin
│   ├── internal/
│   │   ├── audit/                      # bitácora unificada de auditoría
│   │   ├── auth/                       # registro en 3 pasos, login, ARCO, admin de cuentas
│   │   ├── config/                     # carga de variables de entorno
│   │   ├── crypto/                     # Argon2id, HMAC, JWT
│   │   ├── dbmaint/                    # mantenimiento de particiones de tablas
│   │   ├── devices/                    # registro de dispositivos para sync
│   │   ├── domain/                     # tipos de dominio compartidos (User, roles, estados)
│   │   ├── httpjson/ httpproblem/ httputil/  # utilidades HTTP (RFC 7807, decodificación estricta)
│   │   ├── incidents/                  # incidentes de seguridad reportados por el cliente
│   │   ├── levels/                     # secciones y niveles de contenido educativo
│   │   ├── maintenance/                # retención automática (inactividad, cancelación)
│   │   ├── privacy/                    # cancelación de cuenta (ARCO) en una sola transacción
│   │   ├── quiz/                       # banco de preguntas de registro y generación de credenciales
│   │   ├── repository/                 # acceso a datos de la base única
│   │   ├── sync/                       # sincronización de progreso offline
│   │   ├── testdb/                     # esquema Postgres desechable para pruebas de integración
│   │   └── transport/                  # rutas HTTP, middleware, límites de tasa
│   ├── migrations/                     # esquema SQL versionado (golang-migrate)
│   ├── sql/                            # scripts de administración: roles, permisos, seed del primer admin
│   └── main.go                         # punto de entrada del servidor
├── frontend/                           # SPA en React + TypeScript + Vite
│   ├── packages/                       # paquetes locales del workspace npm
│   │   ├── engine/                     # lógica pura de los minijuegos (trivia, memorama, sopa de letras, etc.)
│   │   └── schema/                     # esquemas de validación del contenido educativo de cada plantilla
│   └── src/
│       ├── features/
│       │   ├── auth/                   # registro (wizard de 3 pasos) y login
│       │   ├── admin-accounts/         # alta de staff, borrado, reseteo de contraseña
│       │   ├── admin-quiz-bank/        # CRUD del banco de preguntas de registro
│       │   ├── home/                   # landing mínima tras el login
│       │   └── settings/               # tema, tamaño de texto, filtros de daltonismo, movimiento
│       └── shared/                     # cliente HTTP, componentes de interfaz, esquemas comunes
├── plan/                               # documentos de diseño y planeación técnica
│   ├── 00_Plan_maestro.md              # índice del plan de migración, fases y decisiones abiertas
│   ├── 01_Base_de_datos.md             # esquema de la base de datos
│   ├── 02_Backend.md                   # reparto de paquetes Go
│   ├── 03_Frontend.md                  # historia de los cambios de interfaz (diseño descartado, ver 04)
│   ├── 04_Rediseno_identidad_gustos.md # diseño vigente: cuestionario de gustos, esquema unificado
│   └── Convenciones_de_color_UV.md     # identidad visual institucional (colores, contraste, tipografía)
├── LICENSE                             # licencia del proyecto (Apache 2.0)
└── README.md                           # este archivo
```
