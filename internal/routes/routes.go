package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/maksemen2/avito-shop/config"
	"github.com/maksemen2/avito-shop/internal/handlers"
	"github.com/maksemen2/avito-shop/internal/middleware"
	"go.uber.org/zap"
)

// SetupRoutes настраивает маршруты приложения, устанавливает мидлвари и группирует роутеры.
// Возвращает готовый к запуску роутер.
func SetupRoutes(handler *handlers.RequestsHandler, logger *zap.Logger, corsConfig config.CorsConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.HandleMethodNotAllowed = true

	apiGroup := router.Group("/api")

	apiGroup.Use(middleware.CorsMiddleware(corsConfig))
	apiGroup.Use(middleware.LoggerMiddleware(logger))

	apiGroup.Group("")
	{
		apiGroup.POST("/auth", handler.Authenticate)
	}

	protectedGroup := apiGroup.Group("")
	protectedGroup.Use(middleware.AuthMiddleware(logger, handler.JWTManager))
	{
		protectedGroup.GET("/info", handler.GetInfo)
		protectedGroup.GET("/buy/:item", handler.BuyItem)
		protectedGroup.POST("/sendCoin", handler.SendCoin)
	}

	return router
}
