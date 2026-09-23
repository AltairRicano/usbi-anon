---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Fragmento del componente de registro en el frontend (RegisterPage.tsx) que gestiona el flujo del cuestionario inicial de usuario. Muestra la lógica del primer paso de registro, incluyendo la recolección del flag is_adult por autoreporte del usuario y la validación local de respuestas antes de enviarlas al servidor.

[[pruebas/03_frontend/FE-06_superficie_de_ataque_cliente.md|Documento de análisis FE-06]]

## Funciones

### RegisterPage.handleAnswersSubmit
**Elaboración:** 2026-09-19 | **Actualización:** 2026-09-19

Valida las respuestas del cuestionario utilizando schemas de Zod, verifica la aceptación obligatoria del aviso de privacidad y envía el payload a /auth/register/answers junto con el indicador de mayoría de edad (is_adult) autoreportado.
