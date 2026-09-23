---
tipo: codigo
fecha_elaboracion: 2026-09-21
fecha_actualizacion: 2026-09-21
---
Deriva nickname y password de las respuestas del cuestionario de registro mediante funciones puras, siguiendo el [[Estado_Proyecto/Plan.md|Plan]]. Utiliza dos generadores aleatorios distintos: `math/rand` para el nickname (público) y `crypto/rand` para el password (secreto).

---

## Funciones

### normalizeFragment
**Elaboración:** 2026-09-21 | **Actualización:** 2026-09-21

Reduce una respuesta a minúsculas alfanuméricas puras sin acentos y la recorta para generar los fragmentos del nickname.

### GenerateNicknameCandidates
**Elaboración:** 2026-09-21 | **Actualización:** 2026-09-21

Produce candidatos de nickname combinando respuestas, verificándolos contra colisiones sin acoplarse al repositorio. Invocada por [[backend/internal/auth/service.go.md#Service.RegisterAnswers|Service.RegisterAnswers]].

### generateNicknameCandidate
**Elaboración:** 2026-09-21 | **Actualización:** 2026-09-21

Intenta generar un único candidato aleatorio combinando dos fragmentos con relleno numérico y reintenta si hay colisiones.

### GeneratePassword
**Elaboración:** 2026-09-21 | **Actualización:** 2026-09-21

Produce una contraseña intercalando caracteres de un fragmento legible (derivado por [[backend/internal/quiz/credentials.go.md#passwordFragment|passwordFragment]]) con caracteres aleatorios criptográficamente seguros. Invocada por [[backend/internal/auth/service.go.md#Service.RegisterConfirm|Service.RegisterConfirm]].

### passwordFragment
**Elaboración:** 2026-09-21 | **Actualización:** 2026-09-21

Elige la respuesta más larga, la normaliza con [[backend/internal/quiz/credentials.go.md#normalizeAlnumFull|normalizeAlnumFull]], recorta y capitaliza su primera letra usando [[backend/internal/quiz/credentials.go.md#capitalize|capitalize]].

### normalizeAlnumFull
**Elaboración:** 2026-09-21 | **Actualización:** 2026-09-21

Normaliza una respuesta eliminando caracteres no alfanuméricos y acentos, pero sin recortar la longitud, a diferencia de [[backend/internal/quiz/credentials.go.md#normalizeFragment|normalizeFragment]].

### capitalize
**Elaboración:** 2026-09-21 | **Actualización:** 2026-09-21

Convierte la primera letra del fragmento en mayúscula.

### randomCharsetByte
**Elaboración:** 2026-09-21 | **Actualización:** 2026-09-21

Selecciona un byte seguro y aleatorio a partir de un juego de caracteres predefinido utilizando `crypto/rand`.

### randomDistinctPositions
**Elaboración:** 2026-09-21 | **Actualización:** 2026-09-21

Sortea posiciones distintas para intercalar el fragmento legible entre el relleno de la contraseña.
