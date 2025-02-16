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

func setupTestPurchaseRepository(t *testing.T) (repository.PurchaseRepository, *gorm.DB) {
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
	purchaseRepo := repository.NewPurchaseRepository(db, logger)

	return purchaseRepo, db
}

func TestGetInventoryByUserID_Success(t *testing.T) {
	repo, db := setupTestPurchaseRepository(t)
	ctx := context.Background()

	good := database.Good{
		Type:  "t-shirt",
		Price: 80,
	}
	err := db.Create(&good).Error
	assert.NoError(t, err, "failed to create sample good")

	userID := uint(1)

	purchase1 := database.Purchase{
		UserID: userID,
		GoodID: good.ID,
	}
	purchase2 := database.Purchase{
		UserID: userID,
		GoodID: good.ID,
	}
	err = db.Create(&purchase1).Error
	assert.NoError(t, err, "failed to create purchase1")
	err = db.Create(&purchase2).Error
	assert.NoError(t, err, "failed to create purchase2")

	items, err := repo.GetInventoryByUserID(ctx, userID)
	assert.NoError(t, err, "expected no error retrieving inventory")
	assert.NotNil(t, items, "expected non-nil items slice")
	assert.Equal(t, 1, len(items), "expected one item type in inventory")

	item := items[0]
	assert.Equal(t, "t-shirt", item.Type, "expected good type to match")
	assert.Equal(t, 2, item.Quantity, "expected quantity to be aggregated")
}

func TestGetInventoryByUserID_Empty(t *testing.T) {
	repo, _ := setupTestPurchaseRepository(t)
	ctx := context.Background()

	userID := uint(999)
	items, err := repo.GetInventoryByUserID(ctx, userID)
	assert.NoError(t, err, "expected no error retrieving empty inventory")
	assert.NotNil(t, items, "expected non-nil slice")
	assert.Equal(t, 0, len(items), "expected empty inventory")
}
