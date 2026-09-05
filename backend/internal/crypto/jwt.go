package crypto

import (
	"errors"
	"time"

	"github.com/altair/usbi-anon-backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

type TokenConfig struct {
	Secret       []byte
	AccessExpiry time.Duration
}

type usbiClaims struct {
	UserID       uuid.UUID `json:"user_id"`
	Role         string    `json:"role"`
	TokenVersion int       `json:"token_version"`
	jwt.RegisteredClaims
}

// GenerateToken crea un JWT firmado para un usuario. (Relleno)
func GenerateToken(claims domain.JWTClaims, cfg TokenConfig) (string, error) {
	now := time.Now()

	c := usbiClaims{
		UserID:       claims.UserID,
		Role:         string(claims.Role),
		TokenVersion: claims.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.AccessExpiry)),
			Issuer:    "usbi-backend",
			Subject:   claims.UserID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(cfg.Secret)
}

// ValidateToken analiza y valida un JWT, devolviendo los claims centrales. (Relleno)
func ValidateToken(tokenString string, cfg TokenConfig) (*domain.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &usbiClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return cfg.Secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*usbiClaims); ok && token.Valid {
		return &domain.JWTClaims{
			UserID:       claims.UserID,
			Role:         domain.UserRole(claims.Role),
			TokenVersion: claims.TokenVersion,
		}, nil
	}

	return nil, ErrInvalidToken
}
