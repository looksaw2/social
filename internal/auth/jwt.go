package auth

import "github.com/golang-jwt/jwt/v5"

type JWTAuthenticator struct {
	secret string
	aud    string
	iss    string
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
	return nil, nil
}
