package jwtutil

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"wecheckin-backend/backend/internal/config"
	"wecheckin-backend/backend/pkg/tokenutil"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID, role string) (string, error) {
	secret := config.Cfg.JWT.Secret
	if secret == "" {
		secret = "default-secret-key"
	}
	expire := jwtExpire(role)
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func jwtExpire(role string) time.Duration {
	if config.Cfg.JWT.Expire != "" {
		d, err := time.ParseDuration(config.Cfg.JWT.Expire)
		if err == nil && d > 0 {
			return d
		}
	}
	// fallback to token config
	d, _ := tokenutil.GetTokenConfig(role)
	return d
}

func ParseToken(tokenStr string) (*Claims, error) {
	secret := config.Cfg.JWT.Secret
	if secret == "" {
		secret = "default-secret-key"
	}
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}
