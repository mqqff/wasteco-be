package jwt

import (
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mqqff/wasteco-be/internal/domain/env"
)

type Claims struct {
	jwtlib.RegisteredClaims
	UserID uuid.UUID
}

type JWTIf interface {
	GenerateToken(userID uuid.UUID) (string, error)
	Decode(tokenString string, claims *Claims) error
}

type JWTStruct struct {
	ExpiredTime time.Duration
	SecretKey   string
}

var JWT = getJwt()

func getJwt() JWTIf {
	return &JWTStruct{ExpiredTime: env.AppEnv.JWTExpiredTime, SecretKey: env.AppEnv.JWTSecretKey}
}

func (j *JWTStruct) GenerateToken(userID uuid.UUID) (string, error) {
	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer: env.AppEnv.AppName, Subject: env.AppEnv.AppName,
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(j.ExpiredTime)),
			NotBefore: jwtlib.NewNumericDate(time.Now()), IssuedAt: jwtlib.NewNumericDate(time.Now()),
			ID: uuid.New().String(),
		},
		UserID: userID,
	}
	return jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString([]byte(j.SecretKey))
}

func (j *JWTStruct) Decode(tokenString string, claims *Claims) error {
	token, err := jwtlib.ParseWithClaims(tokenString, claims, func(_ *jwtlib.Token) (any, error) {
		return []byte(j.SecretKey), nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwtlib.ErrSignatureInvalid
	}
	return nil
}
