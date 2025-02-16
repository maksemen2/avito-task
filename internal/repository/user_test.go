package repository_test

import (
	"context"
	"testing"

	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/maksemen2/avito-shop/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func setupTestUserRepository(t *testing.T) (repository.UserRepository, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	if err := db.AutoMigrate(&database.User{}, &database.Purchase{}, &database.Transaction{}, &database.Good{}); err != nil {
		t.Fatalf("failed to migrate User model: %v", err)
	}

	logger := zap.NewNop()
	userRepo := repository.NewUserRepository(db, logger)

	return userRepo, db
}

func TestCreate_Success(t *testing.T) {
	userRepo, _ := setupTestUserRepository(t)
	ctx := context.Background()

	user, err := userRepo.Create(ctx, "testuser", "hashedpassword")
	assert.NoError(t, err, "expected no error creating user")
	assert.Equal(t, user.Username, "testuser", "expected username to match")
	assert.Equal(t, user.PasswordHash, "hashedpassword", "expected password hash to match")

	assert.Equal(t, user.Coins, 1000, "expected user to start with 1000 coins")
}

func TestGetIDByUsername_Success(t *testing.T) {
	userRepo, db := setupTestUserRepository(t)
	ctx := context.Background()

	testUser := &database.User{
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Coins:        100,
	}
	err := db.Create(testUser).Error
	assert.NoError(t, err, "failed to create sample user")

	id, err := userRepo.GetIDByUsername(ctx, "testuser")
	assert.NoError(t, err, "expected no error retrieving user ID")
	assert.Equal(t, testUser.ID, id, "expected returned ID to match inserted user's ID")
}

func TestGetIDByUsername_NotFound(t *testing.T) {
	userRepo, _ := setupTestUserRepository(t)
	ctx := context.Background()

	_, err := userRepo.GetIDByUsername(ctx, "nonexistent")
	assert.Error(t, err, "expected error when user is not found")
	assert.ErrorIs(t, err, repository.ErrUserNotFound, "expected ErrUserNotFound error")
}

func TestGetBalance_Success(t *testing.T) {
	userRepo, db := setupTestUserRepository(t)
	ctx := context.Background()

	testUser := &database.User{
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Coins:        100,
	}

	err := db.Create(testUser).Error
	assert.NoError(t, err, "failed to create sample user")

	balance, err := userRepo.GetBalance(ctx, testUser.ID)
	assert.NoError(t, err, "expected no error retrieving balance")
	assert.Equal(t, testUser.Coins, balance, "expected balance to match inserted user's coins")
}

func TestGetByID_Success(t *testing.T) {
	userRepo, db := setupTestUserRepository(t)
	ctx := context.Background()

	testUser := &database.User{
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Coins:        100,
	}

	err := db.Create(testUser).Error
	assert.NoError(t, err, "failed to create sample user")

	userID := testUser.ID

	user, err := userRepo.GetByID(ctx, userID)
	assert.NoError(t, err, "expected no error retrieving user")
	assert.Equal(t, testUser.Username, user.Username, "expected username to match")
}
