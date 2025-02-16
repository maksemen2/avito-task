package database

import (
	"log"
	"time"

	"github.com/maksemen2/avito-shop/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MustLoad подключается к базе данных и возвращает объект gorm.DB. Завершает работу приложения при ошибке.
func MustLoad(config config.DatabaseConfig) *gorm.DB {
	db, err := gorm.Open(postgres.Open(config.DSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxConnections)
	sqlDB.SetMaxIdleConns(config.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(time.Duration(config.MaxConnectionsLifetimeMinutes) * time.Minute)

	return db
}
