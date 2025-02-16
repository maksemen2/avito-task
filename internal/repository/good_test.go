package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/maksemen2/avito-shop/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func setupTestRepository(t *testing.T) (repository.GoodRepository, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	if err := db.AutoMigrate(&database.User{}, &database.Purchase{}, &database.Transaction{}, &database.Good{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	logger := zap.NewNop()

	repo := repository.NewGoodRepository(db, logger)

	return repo, db
}

func TestGetByName_Success(t *testing.T) {
	repo, db := setupTestRepository(t)

	expectedGood := database.Good{
		Type:  "t-shirt",
		Price: 80,
	}
	err := db.Create(&expectedGood).Error
	assert.NoError(t, err, "failed to create sample good")

	got, err := repo.GetByName(context.Background(), "t-shirt")
	assert.NoError(t, err, "expected no error when retrieving existing good")
	assert.NotNil(t, got, "expected a good to be returned")
	assert.Equal(t, expectedGood.Type, got.Type, "good type should match")
	assert.Equal(t, expectedGood.Price, got.Price, "good price should match")
}

func TestGetByName_NotFound(t *testing.T) {
	repo, _ := setupTestRepository(t)

	got, err := repo.GetByName(context.Background(), "non-existing")
	assert.Nil(t, got, "expected nil result for non-existing good")
	assert.True(t, errors.Is(err, repository.ErrGoodNotFound), "expected error to be ErrGoodNotFound")
}
