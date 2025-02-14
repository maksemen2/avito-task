package auth

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("my_secret")

const jwtExpiryHours = 72

func CreateJwt(userID uint, username string) (string, error) {
	currentTime := time.Now()
	expiresAt := currentTime.Add(time.Hour * jwtExpiryHours)
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"userID":   userID,
			"username": username,
			"exp":      expiresAt.Unix(),
			"iat":      currentTime.Unix(),
			"nbf":      currentTime.Unix(),
		},
	)
	return token.SignedString(jwtKey)
}

func ParseJwt(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
}
