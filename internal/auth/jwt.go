package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/gommon/log"
)

// Claims represents the JWT claims
type Claims struct {
	UserID int    `json:"userId"`
	OpenID string `json:"openId"`
	jwt.RegisteredClaims
}

// Config holds JWT configuration
type Config struct {
	Secret     string
	Issuer     string
	ExpireHour int
}

// DefaultConfig returns default JWT configuration from environment
func DefaultConfig() *Config {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Warn("using defalt secret. change in production.")
		secret = "kuaizu-default-secret-change-in-production"
	}

	return &Config{
		Secret:     secret,
		Issuer:     "kuaizu",
		ExpireHour: 24 * 7, // 7 days
	}
}

// GenerateToken generates a JWT token for a user
func GenerateToken(config *Config, userID int, openID string) (string, int, error) {
	expiresAt := time.Now().Add(time.Duration(config.ExpireHour) * time.Hour)
	expiresIn := config.ExpireHour * 3600 // seconds

	claims := Claims{
		UserID: userID,
		OpenID: openID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.Issuer,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Secret))
	if err != nil {
		return "", 0, fmt.Errorf("sign token: %w", err)
	}

	return tokenString, expiresIn, nil
}

// ParseToken parses and validates a JWT token
func ParseToken(config *Config, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
