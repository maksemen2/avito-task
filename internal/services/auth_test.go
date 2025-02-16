package services_test

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/maksemen2/avito-shop/config"
	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/maksemen2/avito-shop/internal/models"
	"github.com/maksemen2/avito-shop/internal/repository"
	"github.com/maksemen2/avito-shop/internal/services"
	"github.com/maksemen2/avito-shop/pkg/auth"
	"github.com/stretchr/testify/assert"

	gormLogger "gorm.io/gorm/logger"
)

func getMockAuthService(
	t *testing.T,
) services.AuthService {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormLogger.Default.LogMode(gormLogger.Silent)})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	db.AutoMigrate(&database.User{}, &database.Purchase{}, &database.Transaction{}, &database.Good{})

	logger := zap.NewNop()
	holderRepo := repository.NewHolderRepository(db, logger)
	jwtManager := auth.NewJWTManager(config.AuthConfig{JwtKey: "very_secret_key", TokenLifetimeHours: 1})

	return services.NewAuthService(holderRepo, jwtManager, logger)
}

func TestAuthenticate_EmptyCredentials(t *testing.T) {
	authService := getMockAuthService(t)

	req := models.AuthRequest{
		Username: "",
		Password: "",
	}
	resp, err := authService.Authenticate(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, services.ErrUserPassRequired, err)
	assert.Empty(t, resp.Token)
}

func TestAuthenticate_RegisterNewUser(t *testing.T) {
	authService := getMockAuthService(t)

	req := models.AuthRequest{
		Username: "newuser",
		Password: "newpassword",
	}

	_, err := authService.Authenticate(context.Background(), req)
	assert.NoError(t, err)
}

func TestAuthenticate_InvalidPassword(t *testing.T) {
	authService := getMockAuthService(t)

	req := models.AuthRequest{
		Username: "existinguser",
		Password: "goodPassword",
	}

	_, err := authService.Authenticate(context.Background(), req)
	assert.NoError(t, err)

	req = models.AuthRequest{
		Username: "existinguser",
		Password: "badPassword",
	}

	_, err = authService.Authenticate(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, services.ErrAuthFailed, err)
}
