---
tipo: codigo
fecha_elaboracion: 2026-09-19
fecha_actualizacion: 2026-09-21
---

Suite de pruebas unitarias para el algoritmo de generación de credenciales (nicknames y contraseñas) a partir de las respuestas del cuestionario, definido en [[backend/internal/quiz/credentials.go.md|credentials.go]]. Verifica el formato y conjunto de caracteres válidos, la adecuada gestión de colisiones regenerando solo los candidatos afectados, la variabilidad de resultados y la correcta reacción ante respuestas insuficientes o caracteres ambiguos.

[[backend/internal/quiz/credentials.go.md#GenerateNicknameCandidates|GenerateNicknameCandidates]] y [[backend/internal/quiz/credentials.go.md#GeneratePassword|GeneratePassword]], que aquí se prueban, son las funciones que Service.RegisterAnswers invoca para derivar la credencial de una cuenta nueva a partir de sus respuestas al cuestionario de gustos.
