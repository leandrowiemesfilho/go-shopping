package util

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWTUtil interface {
	GenerateToken(userID, email string, expirationHours int) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
}

type jwtUtil struct {
	secret string
	issuer string
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.StandardClaims
}

func NewJWTUtil(secret, issuer string) JWTUtil {
	return &jwtUtil{
		secret: secret,
		issuer: issuer,
	}
}

func (j *jwtUtil) GenerateToken(userID, email string, expirationHours int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    j.issuer,
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

func (j *jwtUtil) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
