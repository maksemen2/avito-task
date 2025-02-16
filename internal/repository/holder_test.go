// language: go
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

func setupTestHolderRepository(t *testing.T) (repository.HolderRepository, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	// Auto migrate necessary models.
	if err := db.AutoMigrate(&database.User{}, &database.Purchase{}, &database.Transaction{}, &database.Good{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	logger := zap.NewNop()
	holderRepo := repository.NewHolderRepository(db, logger)

	return holderRepo, db
}

func TestTransferCoins_Success(t *testing.T) {
	holderRepo, db := setupTestHolderRepository(t)
	ctx := context.Background()

	sender := database.User{Username: "test1", Coins: 100}
	receiver := database.User{Username: "test2", Coins: 50}

	assert.NoError(t, db.Create(&sender).Error, "failed to create sender")
	assert.NoError(t, db.Create(&receiver).Error, "failed to create receiver")

	transferAmount := 30
	err := holderRepo.TransferCoins(ctx, sender.ID, receiver.ID, transferAmount)
	assert.NoError(t, err, "expected successful transfer")

	var updatedSender, updatedReceiver database.User

	assert.NoError(t, db.First(&updatedSender, sender.ID).Error, "failed to fetch sender")
	assert.NoError(t, db.First(&updatedReceiver, receiver.ID).Error, "failed to fetch receiver")
	assert.Equal(t, sender.Coins-transferAmount, updatedSender.Coins, "sender's coins should be deducted")
	assert.Equal(t, receiver.Coins+transferAmount, updatedReceiver.Coins, "receiver's coins should be credited")

	var txRecord database.Transaction
	err = db.Where("from_user_id = ? AND to_user_id = ? AND amount = ?", sender.ID, receiver.ID, transferAmount).First(&txRecord).Error
	assert.NoError(t, err, "expected transaction record creation")
}

func TestTransferCoins_InsufficientFunds(t *testing.T) {
	holderRepo, db := setupTestHolderRepository(t)
	ctx := context.Background()

	sender := database.User{Username: "test1", Coins: 10}
	receiver := database.User{Username: "test2", Coins: 50}

	assert.NoError(t, db.Create(&sender).Error, "failed to create sender")
	assert.NoError(t, db.Create(&receiver).Error, "failed to create receiver")

	transferAmount := 20
	err := holderRepo.TransferCoins(ctx, sender.ID, receiver.ID, transferAmount)
	assert.True(t, errors.Is(err, repository.ErrInsufficientFunds), "expected ErrInsufficientFunds error")

	var updatedSender database.User

	assert.NoError(t, db.First(&updatedSender, sender.ID).Error, "failed to fetch sender")
	assert.Equal(t, sender.Coins, updatedSender.Coins, "sender's coins should remain unchanged")

	var count int64

	assert.NoError(t, db.Model(&database.Transaction{}).Count(&count).Error, "failed to count transactions")
	assert.Equal(t, int64(0), count, "no transaction should be recorded")
}

func TestTransferCoins_ReceiverNotFound(t *testing.T) {
	holderRepo, db := setupTestHolderRepository(t)
	ctx := context.Background()

	sender := database.User{Username: "test1", Coins: 100}
	assert.NoError(t, db.Create(&sender).Error, "failed to create sender")

	nonExistentReceiverID := uint(9999)
	transferAmount := 20
	err := holderRepo.TransferCoins(ctx, sender.ID, nonExistentReceiverID, transferAmount)
	assert.True(t, errors.Is(err, repository.ErrUserNotFound), "expected ErrUserNotFound error")

	var updatedSender database.User

	assert.NoError(t, db.First(&updatedSender, sender.ID).Error, "failed to fetch sender")
	assert.Equal(t, sender.Coins, updatedSender.Coins, "sender's coins should remain unchanged")
}
