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

El proyecto cuenta con paridad funcional de la plataforma educativa gamificada bajo la arquitectura de anonimato por UUID y base de datos única:

- **Backend (Go + PostgreSQL 15)**: esquema unificado de 26+ tablas (migraciones `0001` a `0006`), registro en 3 pasos mediante cuestionario de gustos no sensibles, login sin correo, generación determinista de nickname y password de 12 caracteres, roles reducidos estrictamente a `player` y `admin`, catálogo de contenido educativo (secciones y niveles) con validación estricta en servidor de 7 plantillas, rotación de temporadas con purga real de espacio preservando la XP ganada, lectura de auditoría con auto-auditoría, incidentes de seguridad, CRUD de insignias, enlaces de interés comunitarios, buzón de sugerencias, y registro y revocación de dispositivos.
- **Frontend (React 18 + TypeScript + Vite + npm workspaces)**:
  - Experiencia del jugador completa: Dashboard interactivo, catálogo de secciones, reproducción de niveles con 7 minijuegos interactivos (Trivia, Memorama, Fake News, Sopa de letras, Rompecabezas, Crucigrama, Serpientes y Escaleras) con integración de escenas Phaser, racha diaria, insignias y progreso persistente en perfil.
  - Maker local (`/maker`) para diseñar, probar y exportar/importar niveles en formato JSON sin privilegios de administrador.
  - Paneles administrativos: gestión de staff, banco de preguntas de registro, administración de contenido educativo (con previsualización, archivo y purga irreversible con doble confirmación), insignias, seguridad/auditoría y comunidad.
  - Personalización y accesibilidad (`/settings` pública): temas claro/oscuro, escala tipográfica, filtros CSS de daltonismo compatibles entre navegadores (Chrome/Firefox) y reducción de movimiento.
  - Navegación estandarizada (`HomeButton` y `LinkButton`) y carrusel comunitario animado con verificación de contraste WCAG 2.2.
- **En curso**: fase de maduración hacia producción, descrita en [`plan/00_Plan_de_maduracion.md`](plan/00_Plan_de_maduracion.md). Cubre la validación real del resultado de cada nivel (criterio de victoria por plantilla y verificación en servidor), la redacción e integración del aviso de privacidad, el despliegue reproducible en servidor propio, y la documentación técnica y legal del sistema.
- **Deuda conocida**: paginación en catálogos extensos de contenido, pulido del formulario de administración de comunidad, y borrado duro directo de la cuenta de jugador.

## Arquitectura de datos

A diferencia de un sistema tradicional, **una sola base de datos** contiene tanto las credenciales de acceso como el progreso y contenido educativo, indexados exclusivamente por UUID. Ninguna tabla contiene nombre, correo, teléfono ni cualquier otro dato de contacto: el nickname que identifica a cada cuenta se genera a partir de fragmentos de las respuestas al cuestionario de registro, no de un dato personal.

Retirar contenido educativo entre temporadas (rotación de niveles) preserva siempre la experiencia (XP) y los contadores de progreso ya ganados por cada jugador (`account_retired_progress` y `experience_history.level_id SET NULL`), aunque el nivel original se purgue físicamente para liberar almacenamiento en el servidor (~20 GB).

## Identidad visual

El proyecto se presenta como producto de la Universidad Veracruzana / USBI y sigue las convenciones institucionales de color, contraste y accesibilidad documentadas en [`plan/Convenciones_de_color_UV.md`](plan/Convenciones_de_color_UV.md).

## Estructura del proyecto

```text
.
├── backend/                            # API en Go (chi router, PostgreSQL)
│   ├── cmd/
│   │   └── hash_password/              # binario standalone para el bootstrap del primer admin
│   ├── internal/
│   │   ├── audit/                      # utilidades de registro de auditoría
│   │   ├── auditlog/                   # lectura de bitácora de auditoría y auto-auditoría
│   │   ├── auth/                       # registro en 3 pasos, login, gestión de credenciales y staff
│   │   ├── badges/                     # CRUD de insignias y asignación
│   │   ├── config/                     # carga y validación de variables de entorno
│   │   ├── crypto/                     # Argon2id, HMAC, JWT
│   │   ├── dbmaint/                    # mantenimiento de particiones de tablas
│   │   ├── devices/                    # alta automática y revocación de dispositivos
│   │   ├── domain/                     # tipos de dominio compartidos (User, roles 'player' y 'admin')
│   │   ├── httpjson/ httpproblem/ httputil/  # utilidades HTTP (RFC 7807, decodificación estricta)
│   │   ├── incidents/                  # incidentes de seguridad reportados y gestión
│   │   ├── interestlinks/              # enlaces de interés comunitarios
│   │   ├── levels/                     # secciones y niveles, validadores de 7 plantillas, purga real
│   │   ├── maintenance/                # retención automática y limpieza periódica
│   │   ├── privacy/                    # cancelación de cuentas en transacción atómica
│   │   ├── quiz/                       # banco de preguntas de registro y derivación de nickname/password
│   │   ├── repository/                 # capa de acceso a datos sobre la base unificada
│   │   ├── suggestions/                # buzón y gestión de sugerencias
│   │   ├── sync/                       # sincronización e historial offline
│   │   ├── testdb/                     # esquema Postgres desechable para pruebas de integración
│   │   └── transport/                  # enrutador Chi, middlewares de seguridad, límites de tasa
│   ├── migrations/                     # esquema SQL versionado (0001 a 0006)
│   ├── sql/                            # scripts de roles (00_roles_unificado.sql) y seed del admin
│   └── main.go                         # punto de entrada del servidor Go
├── frontend/                           # SPA en React + TypeScript + Vite
│   ├── packages/                       # paquetes locales del workspace npm
│   │   ├── engine/                     # lógica pura de los 7 minijuegos (sin DOM/React)
│   │   └── schema/                     # esquemas Zod de validación de contenido por plantilla
│   └── src/
│       ├── features/
│       │   ├── admin-accounts/         # alta y administración de cuentas staff, reseteo de contraseñas
│       │   ├── admin-badges/           # administración y catálogo de insignias
│       │   ├── admin-community/        # gestión de tarjetas del carrusel y sugerencias
│       │   ├── admin-quiz-bank/        # CRUD del banco de preguntas de registro (guard mínimo 4)
│       │   ├── admin-security/         # visualización de auditoría e incidentes de seguridad
│       │   ├── auth/                   # wizard de registro en 3 pasos y login
│       │   ├── content/                # panel admin de contenido, editor por plantilla, archivo y purga
│       │   ├── dashboard/              # catálogo de secciones, racha, XP, carrusel de enlaces de interés
│       │   ├── games/                  # vistas y componentes interactivos de minijuegos (Phaser)
│       │   ├── maker/                  # maker local: diseño de niveles en localStorage y exportación JSON
│       │   ├── offline-processes/      # monitoreo de dispositivos y registros de sincronización
│       │   ├── profile/                # perfil de jugador, insignias obtenidas, XP y diálogo de baja
│       │   └── settings/               # personalización, temas, escala de texto, daltonismo cross-browser
│       └── shared/                     # cliente API, botones de navegación estandarizados, UI base
├── plan/                               # documentos de diseño y planeación técnica
│   ├── 00_Plan_de_maduracion.md        # plan vigente: veracidad de resultados, avisos de
│   │                                   #   privacidad, despliegue y documentación
│   └── Convenciones_de_color_UV.md     # guía de diseño institucional UV
├── LICENSE                             # licencia del proyecto (Apache 2.0)
└── README.md                           # este archivo
```
