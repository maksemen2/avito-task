package main

import (
	"log"

	"github.com/maksemen2/avito-task/config"
	"github.com/maksemen2/avito-task/internal/database"
)

func main() {
	cfg := config.MustLoad()
	db := database.MustLoad(cfg.Database.DSN())

	// Автоматическая миграция для моделей
	if err := db.AutoMigrate(&database.User{}, &database.Purchase{}, &database.Transaction{}); err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}

	log.Println("Миграция успешно завершена")
}
