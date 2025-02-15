package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable TimeZone=UTC",
		c.Host, c.Port, c.User, c.DBName, c.Password)
}

type AuthConfig struct {
	JwtKey             string
	TokenLifetimeHours int
}

type Config struct {
	Database DatabaseConfig
	Auth     AuthConfig
}

func MustLoad() *Config {
	err := godotenv.Load()

	if err != nil {
		err = godotenv.Load(".env.dist")
		if err != nil {
			log.Fatalf("Ошибка загрузки файлов окружения: %v", err)
		}
	}

	tokenLifetimeStr := os.Getenv("TOKEN_LIFETIME_HOURS")
	tokenLifetime, err := strconv.Atoi(tokenLifetimeStr)

	if err != nil {
		tokenLifetime = 24
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     os.Getenv("DATABASE_HOST"),
			Port:     os.Getenv("DATABASE_PORT"),
			User:     os.Getenv("DATABASE_USERNAME"),
			Password: os.Getenv("DATABASE_PASSWORD"),
			DBName:   os.Getenv("DATABASE_NAME"),
		},
		Auth: AuthConfig{
			JwtKey:             os.Getenv("JWT_SECRET"),
			TokenLifetimeHours: tokenLifetime,
		},
	}
}
