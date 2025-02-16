package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/maksemen2/avito-shop/internal/models"
	"github.com/maksemen2/avito-shop/pkg/auth"
	"go.uber.org/zap"
)

func AuthMiddleware(logger *zap.Logger, jwtManager *auth.JWTManager) gin.HandlerFunc {
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
			logger.Warn("Invalid token structure",
				zap.String("ip addr", c.ClientIP()),
				zap.String("token", tokenStr))
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Errors: models.ErrUnauthorized})

			return
		}

		userIDFloat, userIDExists := claims[auth.UserIDKey].(float64)
		username, usernameExists := claims[auth.UsernameKey].(string)

		if !userIDExists || !usernameExists {
			logger.Warn("Malformed token",
				zap.Any("claims", claims),
				zap.String("ip addr", c.ClientIP()))
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Errors: models.ErrUnauthorized})

			return
		}

		c.Set(auth.UserIDKey, uint(userIDFloat))
		c.Set(auth.UsernameKey, username)
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

func GetUsername(c *gin.Context) (string, bool) {
	username, ok := c.Get(auth.UsernameKey)
	if !ok {
		return "", false
	}

	return username.(string), true
}
