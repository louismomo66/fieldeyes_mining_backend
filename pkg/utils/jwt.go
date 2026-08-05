package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtSecret holds the signing key. It must be set via SetJWTSecret before any
// token is generated or validated. There is intentionally no hardcoded fallback.
var jwtSecret []byte

// SetJWTSecret sets the JWT secret key. Must be called at startup with a value
// loaded from the environment. Panics if secret is empty.
func SetJWTSecret(secret string) {
	if secret == "" {
		panic("JWT_SECRET environment variable must be set and non-empty")
	}
	if len(secret) < 32 {
		panic("JWT_SECRET must be at least 32 characters long")
	}
	jwtSecret = []byte(secret)
}

// Claims represents JWT claims.
// ChainRole is the supply-chain role (operator/transporter/exporter/inspector).
// It is omitempty and optional: tokens issued before this feature simply have an
// empty ChainRole, which the app treats as legacy full access.
type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ChainRole string `json:"chain_role,omitempty"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new JWT token (no chain role — kept for compatibility).
func GenerateJWT(userID, email, role string) (string, error) {
	return GenerateJWTWithChainRole(userID, email, role, "")
}

// GenerateJWTWithChainRole creates a token that also carries the supply-chain role.
func GenerateJWTWithChainRole(userID, email, role, chainRole string) (string, error) {
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		ChainRole: chainRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWT validates a JWT token
func ValidateJWT(tokenString string) (*Claims, error) {
	if len(jwtSecret) == 0 {
		return nil, errors.New("JWT secret not initialised")
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Reject tokens that don't use HMAC — prevents algorithm confusion attacks
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateToken is an alias for GenerateJWT for backward compatibility
func GenerateToken(userID, email, role string) (string, error) {
	return GenerateJWT(userID, email, role)
}
