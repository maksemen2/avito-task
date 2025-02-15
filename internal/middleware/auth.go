package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/maksemen2/avito-shop/internal/auth"
	"github.com/maksemen2/avito-shop/internal/models"
)

func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Errors: models.ErrUnauthorized})
			return
		}
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		t, err := jwtManager.ParseToken(tokenStr)
		if err != nil || !t.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Errors: models.ErrUnauthorized})
			return
		}

		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Errors: models.ErrUnauthorized})
			return
		}

		userIDFloat, ok := claims[auth.UserIDKey].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Errors: models.ErrUnauthorized})
			return
		}

		c.Set(auth.UserIDKey, uint(userIDFloat))
		c.Next()
	}
}

func GetUserID(c *gin.Context) (uint, bool) {
	userID, ok := c.Get(auth.UserIDKey)
	if !ok {
		return 0, false
	}

	return userID.(uint), true
}
