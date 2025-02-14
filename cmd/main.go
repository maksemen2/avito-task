package main

import (
	"github.com/maksemen2/avito-task/internal/database"
	"github.com/maksemen2/avito-task/internal/handlers"
	"github.com/maksemen2/avito-task/internal/routes"
)

func main() {
	db := database.MustLoad("")
	requestsHandler := handlers.NewRequestsHandler(db)
	router := routes.SetupRoutes(requestsHandler)
	router.Run(":8080")
}
