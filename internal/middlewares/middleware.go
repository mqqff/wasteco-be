package middlewares

import "github.com/mqqff/wasteco-be/pkg/jwt"

type Middleware struct {
	jwt jwt.JWTIf
}
