// Package badges implementa el CRUD administrativo del catálogo de insignias.
// La concesión automática por umbral de XP y la consulta del jugador se gestionan en internal/levels
// y repository/badge_queries.go respectivamente.
//
// Solo AdminService: no existe un PlayerService aquí ya que la lectura del jugador se gestiona en levels.
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

// iconKeyPattern valida un slug libre en lugar de un catálogo cerrado, permitiendo al frontend
// mapear iconos dinámicamente y usar un icono por defecto si no reconoce la clave.
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
