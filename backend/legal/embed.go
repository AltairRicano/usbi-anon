// Package legaltext empotra los textos de los avisos de privacidad en el
// binario (M2.4 punto 1 del plan de maduración): el texto legal vive
// versionado en el repositorio, no en una constante del cliente ni en un
// archivo cargado aparte al servidor.
//
// TODO(M4.3): sustituir aviso_simplificado_v1.json y aviso_integral_v1.json
// por la redacción legal definitiva, y CurrentVersion por "v1.0" — ambos
// archivos son, a propósito, un marcador de posición explícito hasta que se
// cierren las decisiones institucionales pendientes (responsable del
// tratamiento, nombre/dominio, ubicación del centro de datos).
package legaltext

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

//go:embed aviso_simplificado_v1.json
var simplifiedRaw []byte

//go:embed aviso_integral_v1.json
var integralRaw []byte

// CurrentVersion identifica el texto vigente. No se despliega a producción
// con este valor (D-06 asume "v1.0-preliminar" desaparece antes del primer
// despliegue) — "v0-provisional" lo deja imposible de confundir con una
// versión real.
const CurrentVersion = "v0-provisional"

// EffectiveDate acompaña a CurrentVersion. Formato ISO 8601 (fecha sola).
const EffectiveDate = "2026-09-17"

// Section es una sección de aviso: encabezado + párrafos. JSON estructurado
// en vez de Markdown/HTML (M2.4 punto 3): sin intérprete de Markdown en el
// bundle, sin dangerouslySetInnerHTML, y con encabezados reales para lector
// de pantalla en el frontend.
type Section struct {
	Heading    string   `json:"heading"`
	Paragraphs []string `json:"paragraphs"`
}

// Notice es el aviso vigente completo, ya parseado y con su checksum.
type Notice struct {
	Version       string
	EffectiveDate string
	Simplified    []Section
	Full          []Section
	Checksum      string
}

type noticeFile struct {
	Sections []Section `json:"sections"`
}

var current = mustBuildCurrent()

func mustBuildCurrent() Notice {
	var simplified, full noticeFile
	if err := json.Unmarshal(simplifiedRaw, &simplified); err != nil {
		panic("legaltext: aviso_simplificado_v1.json malformado: " + err.Error())
	}
	if err := json.Unmarshal(integralRaw, &full); err != nil {
		panic("legaltext: aviso_integral_v1.json malformado: " + err.Error())
	}

	sum := sha256.New()
	sum.Write(simplifiedRaw)
	sum.Write(integralRaw)

	return Notice{
		Version:       CurrentVersion,
		EffectiveDate: EffectiveDate,
		Simplified:    simplified.Sections,
		Full:          full.Sections,
		Checksum:      hex.EncodeToString(sum.Sum(nil)),
	}
}

// Current devuelve el aviso vigente.
func Current() Notice {
	return current
}

// VerifyVersion confirma que `version` es la vigente. Se usa al registrar una
// cuenta (M2.4 punto 4): un cliente con caché vieja se rechaza con 409 en vez
// de sellar la aceptación de una versión que el servidor ya no sirve.
func VerifyVersion(version string) bool {
	return version != "" && version == CurrentVersion
}

// SealPayload construye el material que se firma con HMAC para sellar la
// aceptación del aviso vigente por una cuenta. Ligado no solo a la etiqueta
// de versión sino al checksum del texto realmente servido (M2.4 punto 4):
// así un cliente modificado no puede declarar "acepté v9.0" sobre un
// contenido que el servidor nunca sirvió con ese checksum.
func SealPayload(accountID uuid.UUID, acceptedAt time.Time) []byte {
	n := Current()
	return []byte(accountID.String() + "|" + n.Version + "|" + n.Checksum + "|" + acceptedAt.Format(time.RFC3339Nano))
}
