package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/maksemen2/avito-task/internal/handlers"
	"github.com/maksemen2/avito-task/internal/middleware"
)

func SetupRoutes(handler *handlers.RequestsHandler) *gin.Engine {
	router := gin.Default()

	apiGroup := router.Group("/api")

	apiGroup.Group("/auth", handler.Authenticate)

	protectedGroup := apiGroup.Group("")
	protectedGroup.Use(middleware.AuthMiddleware(handler.JWTManager))
	{
		protectedGroup.GET("/info", handler.GetInfo)
		protectedGroup.GET("/buy/:item", handler.BuyItem)
		protectedGroup.POST("/sendCoin", handler.SendCoin)
	}

	return router
}
