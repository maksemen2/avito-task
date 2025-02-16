package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type DatabaseConfig struct {
	Host                          string
	Port                          string
	User                          string
	Password                      string
	DBName                        string
	MaxConnections                int
	MaxIdleConnections            int
	MaxConnectionsLifetimeMinutes int
}

// DSN возвращает строку подключения к базе данных для GORM.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable TimeZone=UTC",
		c.Host, c.Port, c.User, c.DBName, c.Password)
}

type AuthConfig struct {
	JwtKey             string
	TokenLifetimeHours int
}

type CorsConfig struct {
	AllowedOrigins   string
	AllowedMethods   string
	AllowedHeaders   string
	AllowCredientals string
	MaxAge           string
}

type LoggerConfig struct {
	Level    string
	FilePath string
}

type Config struct {
	Database DatabaseConfig
	Auth     AuthConfig
	Cors     CorsConfig
	Logger   LoggerConfig
}

// LoadEnv загружает переменные окружения из файла .env или .env.dist.
func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load(".env.dist"); err != nil {
			return fmt.Errorf("ошибка загрузки файлов окружения: %v", err)
		}
	}

	return nil
}

func LoadDatabaseConfig() (DatabaseConfig, error) {
	maxConnections, err := strconv.Atoi(os.Getenv("DATABASE_MAX_CONNECTIONS"))
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("ошибка преобразования DATABASE_MAX_CONNECTIONS: %v", err)
	}

	maxIdleConnections, err := strconv.Atoi(os.Getenv("DATABASE_MAX_IDLE_CONNECTIONS"))
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("ошибка преобразования DATABASE_MAX_IDLE_CONNECTIONS: %v", err)
	}

	maxConnectionsLifetimeMinutes, err := strconv.Atoi(os.Getenv("DATABASE_MAX_CONNECTIONS_LIFETIME_MINUTES"))
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("ошибка преобразования DATABASE_MAX_CONNECTIONS_LIFETIME_MINUTES: %v", err)
	}

	return DatabaseConfig{
		Host:                          os.Getenv("DATABASE_HOST"),
		Port:                          os.Getenv("DATABASE_PORT"),
		User:                          os.Getenv("DATABASE_USERNAME"),
		Password:                      os.Getenv("DATABASE_PASSWORD"),
		DBName:                        os.Getenv("DATABASE_NAME"),
		MaxConnections:                maxConnections,
		MaxIdleConnections:            maxIdleConnections,
		MaxConnectionsLifetimeMinutes: maxConnectionsLifetimeMinutes,
	}, nil
}

func LoadAuthConfig() (AuthConfig, error) {
	tokenLifetimeStr := os.Getenv("TOKEN_LIFETIME_HOURS")

	tokenLifetime, err := strconv.Atoi(tokenLifetimeStr)
	if err != nil {
		tokenLifetime = 24
	}

	return AuthConfig{
		JwtKey:             os.Getenv("JWT_SECRET"),
		TokenLifetimeHours: tokenLifetime,
	}, nil
}

func LoadCorsConfig() CorsConfig {
	return CorsConfig{
		AllowedOrigins:   os.Getenv("CORS_ALLOWED_ORIGINS"),
		AllowedMethods:   os.Getenv("CORS_ALLOWED_METHODS"),
		AllowedHeaders:   os.Getenv("CORS_ALLOWED_HEADERS"),
		AllowCredientals: os.Getenv("CORS_ALLOW_CREDENTIALS"),
		MaxAge:           os.Getenv("CORS_MAX_AGE"),
	}
}

func LoadLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:    os.Getenv("LOG_LEVEL"),
		FilePath: os.Getenv("LOG_FILE"),
	}
}

// MustLoad загружает конфигурацию из переменных окружения.
// Если переменные не указаны, используются значения по умолчанию.
// В случае ошибки завершает работу программы.
// LoadDatabaseConfig загружает конфигурацию базы данных из переменных окружения.
func MustLoad() *Config {
	if err := LoadEnv(); err != nil {
		log.Fatalf("Ошибка загрузки окружения: %v", err)
	}

	dbConfig, err := LoadDatabaseConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации базы данных: %v", err)
	}

	authConfig, err := LoadAuthConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации аутентификации: %v", err)
	}

	return &Config{
		Database: dbConfig,
		Auth:     authConfig,
		Cors:     LoadCorsConfig(),
		Logger:   LoadLoggerConfig(),
	}
}
