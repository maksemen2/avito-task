package main

import (
	"github.com/maksemen2/avito-task/config"
	"github.com/maksemen2/avito-task/internal/auth"
	"github.com/maksemen2/avito-task/internal/database"
	"github.com/maksemen2/avito-task/internal/handlers"
	"github.com/maksemen2/avito-task/internal/routes"
)

func main() {
	config := config.MustLoad()
	jwtManager := auth.NewJWTManager(config.Auth)
	db := database.MustLoad(config.Database.DSN())
	requestsHandler := handlers.NewRequestsHandler(db, jwtManager)
	router := routes.SetupRoutes(requestsHandler)
	router.Run(":8080")
}
