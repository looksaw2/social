package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
	secret string //token
	aud    string //期望接受的目标
	iss    string //签发的目标
}

// 新建一个JWTAuthenticator
func NewJWTAuthenticator(secret string, aud string, iss string) *JWTAuthenticator {
	return &JWTAuthenticator{
		secret: secret,
		aud:    aud,
		iss:    iss,
	}
}

// 生成一个Token
func (a *JWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	//通过claims得到Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//得到Tokenstring
	tokenString, err := token.SignedString([]byte(a.secret))

	if err != nil {
		return "", nil
	}
	return tokenString, nil
}

// 验证Token
func (a *JWTAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(a.secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(a.aud),
		jwt.WithIssuer(a.aud),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}
