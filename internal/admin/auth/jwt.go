package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/gommon/log"
)

// AdminClaims represents the admin JWT claims
type AdminClaims struct {
	AdminID  int    `json:"adminId"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AdminConfig holds admin JWT configuration
type AdminConfig struct {
	Secret     string
	Issuer     string
	ExpireHour int
}

// DefaultAdminConfig returns default admin JWT configuration from environment
func DefaultAdminConfig() *AdminConfig {
	secret := os.Getenv("ADMIN_JWT_SECRET")
	if secret == "" {
		log.Warn("using default admin secret. change in production.")
		secret = "kuaizu-admin-default-secret-change-in-production"
	}

	return &AdminConfig{
		Secret:     secret,
		Issuer:     "kuaizu-admin",
		ExpireHour: 8,
	}
}

// GenerateAdminToken generates a JWT token for an admin user
func GenerateAdminToken(config *AdminConfig, adminID int, username string) (string, int, error) {
	expiresAt := time.Now().Add(time.Duration(config.ExpireHour) * time.Hour)
	expiresIn := config.ExpireHour * 3600

	claims := AdminClaims{
		AdminID:  adminID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.Issuer,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Secret))
	if err != nil {
		return "", 0, fmt.Errorf("sign admin token: %w", err)
	}

	return tokenString, expiresIn, nil
}

// ParseAdminToken parses and validates an admin JWT token
func ParseAdminToken(config *AdminConfig, tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse admin token: %w", err)
	}

	if claims, ok := token.Claims.(*AdminClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid admin token")
}
