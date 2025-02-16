package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maksemen2/avito-shop/config"
	"github.com/maksemen2/avito-shop/internal/auth"
	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/maksemen2/avito-shop/internal/handlers"
	"github.com/maksemen2/avito-shop/internal/routes"
	"github.com/maksemen2/avito-shop/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	config := config.MustLoad()
	logger := logger.MustLoad(config.Logger)
	jwtManager := auth.NewJWTManager(config.Auth)
	db := database.MustLoad(config.Database)
	requestsHandler := handlers.NewRequestsHandler(db, jwtManager, logger)
	router := routes.SetupRoutes(requestsHandler, logger, config.Cors)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to run server", zap.Error(err))
		}
	}()

	logger.Info("Server is running on http://localhost:8080")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Гарантируем очистку ресурсов

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Shutdown forced", zap.Error(err))
	}

	logger.Info("Exit")
}
