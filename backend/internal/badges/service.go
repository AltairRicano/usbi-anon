// Package badges implementa el CRUD administrativo del catálogo de
// insignias (B3, estado_proyecto.md 2026-09-09 — hueco de producto abierto
// desde 2026-09-02: "el equipo de USBI podrá gestionar insignias sin
// depender de un ingeniero"). Antes de este paquete, badges solo tenía
// lectura (levels.GetProfileProgress → repository.ListUserBadges) y la
// concesión automática por umbral de XP (repository.AwardEligibleBadges) —
// ninguna de las dos toca este código, siguen viviendo en internal/levels y
// internal/repository/badge_queries.go respectivamente.
//
// Solo AdminService: no existe un PlayerService aquí porque la lectura del
// jugador ya está resuelta en otro paquete y no cambia con B3.
package badges

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrValidation     = errors.New("validation error")
	ErrNotFound       = errors.New("not found")
	ErrBadgeHasHolder = errors.New("badge already earned by at least one account")
)

// iconKeyPattern es deliberadamente un slug libre, no un catálogo cerrado
// (decisión D2, estado_proyecto.md 2026-09-09): el frontend todavía no tiene
// un mapa de assets por icon_key, así que fijar aquí una lista cerrada
// inventaría un contrato por adelantado y obligaría a tocar el backend cada
// vez que se agregue un icono nuevo. El frontend cae a un icono por defecto
// si no reconoce la clave.
var iconKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{2,39}$`)

const nameMaxLen = 100

func validateBadgeInput(name, iconKey string, xpThreshold int32) error {
	if name == "" || len(name) > nameMaxLen {
		return ErrValidation
	}
	if xpThreshold < 0 {
		return ErrValidation
	}
	if !iconKeyPattern.MatchString(iconKey) {
		return ErrValidation
	}
	return nil
}

func trimBadgeInput(name, iconKey string) (string, string) {
	return strings.TrimSpace(name), strings.TrimSpace(iconKey)
}
