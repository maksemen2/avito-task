package auth

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/maksemen2/avito-shop/config"
)

type JWTManager struct {
	signingKey    []byte
	tokenDuration time.Duration
}

const UserIDKey = "userID"

// NewJWTManager создает новый экземпляр JWTManager.
// signingKey - ключ для подписи токена.
// tokenDuration - длительность жизни токена в часах.
func NewJWTManager(config config.AuthConfig) *JWTManager {
	return &JWTManager{
		signingKey:    []byte(config.JwtKey),
		tokenDuration: time.Duration(config.TokenLifetimeHours) * time.Hour,
	}
}

func (m *JWTManager) GenerateToken(userID uint, username string) (string, error) {
	now := time.Now()
	expireTime := now.Add(m.tokenDuration)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		UserIDKey:  userID,
		"username": username,
		"exp":      expireTime.Unix(),
		"iat":      now.Unix(),
		"nbf":      now.Unix(),
	})

	return token.SignedString(m.signingKey)
}

func (m *JWTManager) ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return m.signingKey, nil
	})
}
