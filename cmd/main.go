package main

import (
	"github.com/maksemen2/avito-shop/config"
	"github.com/maksemen2/avito-shop/internal/auth"
	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/maksemen2/avito-shop/internal/handlers"
	"github.com/maksemen2/avito-shop/internal/routes"
)

func main() {
	config := config.MustLoad()
	jwtManager := auth.NewJWTManager(config.Auth.JwtKey, config.Auth.TokenLifetimeHours)
	db := database.MustLoad(config.Database.DSN())
	requestsHandler := handlers.NewRequestsHandler(db, jwtManager)
	router := routes.SetupRoutes(requestsHandler)
	router.Run(":8080")
}
