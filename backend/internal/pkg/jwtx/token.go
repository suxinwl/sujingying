package jwtx

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	Typ    string `json:"typ"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID uint, role, secret string, minutes int) (string, error) {
	return generateToken(userID, role, secret, "access", time.Duration(minutes)*time.Minute)
}

func GenerateRefreshToken(userID uint, role, secret string, hours int) (string, error) {
	return generateToken(userID, role, secret, "refresh", time.Duration(hours)*time.Hour)
}

func generateToken(userID uint, role, secret, typ string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		Typ:    typ,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func Parse(tokenStr, secret string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if c, ok := t.Claims.(*Claims); ok && t.Valid {
		return c, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
