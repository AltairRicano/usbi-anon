# USBI-Anon

Documentación de diseño para una variante de la plataforma educativa
gamificada USBI (Universidad Veracruzana) en la que la identidad de cada
usuario queda representada únicamente por un identificador UUID en el sistema
principal, separando la información de contacto y credenciales en una base de
datos independiente.

El sistema conserva las mismas capacidades funcionales que la propuesta
original: catálogo de secciones y niveles, minijuegos, experiencia (XP),
insignias, rachas, progreso persistente entre dispositivos y panel de
administración. El cambio de fondo es de arquitectura de datos y de modelo de
alojamiento, no de alcance funcional.

Licencia: Apache 2.0. Consulta [LICENSE](LICENSE) cuando se incorpore junto
con el código.

## Estado actual

Este repositorio contiene la documentación de arquitectura, el plan técnico de
la migración, las convenciones de identidad visual y los **esquemas SQL de las
dos bases de datos**, ya escritos y verificados contra PostgreSQL 15. No incluye
todavía el código de aplicación (backend ni frontend): esa implementación parte
de una plataforma hermana ya existente y se incorporará siguiendo las fases
descritas en [`plan/00_Plan_maestro.md`](plan/00_Plan_maestro.md).

## Arquitectura de datos

El diseño divide la persistencia en dos bases de datos con distinto nivel de
sensibilidad:

- **Base de datos privada de identidad**: guarda exclusivamente correo
  electrónico, contraseña (con hash) y el UUID asociado a cada persona
  usuaria. Es la única pieza del sistema que constituye un dato personal
  identificable de forma directa.
- **Base de datos principal**: guarda todo el progreso, contenido educativo,
  insignias, rachas y actividad administrativa, indexado por UUID. No
  contiene nombre, correo, teléfono ni ningún otro dato de contacto.

Esta separación permite que la base principal quede alojada en un servidor
institucional sin necesidad de recurrir a un proveedor de hospedaje externo
para la información sensible del proyecto.

## Identidad visual

El proyecto se presenta como producto de la Universidad Veracruzana / USBI y
por lo tanto sigue las convenciones institucionales de color, contraste y
accesibilidad documentadas en [`plan/Convenciones_de_color_UV.md`](plan/Convenciones_de_color_UV.md).

## Estructura del proyecto

```text
.
├── backend/
│   ├── migrations/
│   │   ├── identity/                  # esquema de la base de identidad (credenciales)
│   │   └── main/                      # esquema de la base principal (progreso y contenido)
│   └── sql/                           # scripts de administración: roles y permisos
├── plan/                              # documentos de diseño y planeación técnica
│   ├── 00_Plan_maestro.md             # índice del plan de migración, fases y decisiones abiertas
│   ├── 01_Base_de_datos.md            # esquema de las dos bases y scripts SQL a producir
│   ├── 02_Backend.md                  # reparto de servicios entre ambas bases
│   ├── 03_Frontend.md                 # alcance de los cambios de interfaz
│   └── Convenciones_de_color_UV.md    # identidad visual institucional (colores, contraste, tipografía)
├── LICENSE                            # licencia del proyecto (Apache 2.0)
└── README.md                          # este archivo
```

La estructura se ampliará con `backend/`, `frontend/` y `docs_referencia/`
cuando se incorpore la implementación técnica y la documentación legal
completa del proyecto.
