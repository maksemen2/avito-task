package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/maksemen2/avito-shop/config"
)

func CorsMiddleware(config config.CorsConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Credentials", config.AllowCredientals)
		c.Writer.Header().Set("Access-Control-Max-Age", config.MaxAge)
		c.Writer.Header().Set("Access-Control-Allow-Origin", config.AllowedOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Methods", config.AllowedMethods)
		c.Writer.Header().Set("Access-Control-Allow-Headers", config.AllowedHeaders)

		c.Next()
	}
}
