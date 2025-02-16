package services_test

import (
	"context"
	"testing"

	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/maksemen2/avito-shop/internal/models"
	"github.com/maksemen2/avito-shop/internal/repository"
	"github.com/maksemen2/avito-shop/internal/services"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func getMockTransferService(t *testing.T) (services.TransferService, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormLogger.Default.LogMode(gormLogger.Silent)})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	db.AutoMigrate(&database.User{}, &database.Purchase{}, &database.Transaction{}, &database.Good{})

	logger := zap.NewNop()
	holderRepo := repository.NewHolderRepository(db, logger)

	return services.NewTransferService(holderRepo, logger), db
}

func TestTransferCoins_Success(t *testing.T) {
	srv, db := getMockTransferService(t)

	sender := database.User{
		Username:     "test1",
		PasswordHash: "test1",
	}
	receiver := database.User{
		Username:     "test2",
		PasswordHash: "test2",
	}

	assert.NoError(t, db.Create(&sender).Error, "failed to create user1")
	assert.NoError(t, db.Create(&receiver).Error, "failed to create user2")

	req := models.SendCoinRequest{
		ToUser: receiver.Username,
		Amount: 100,
	}

	err := srv.SendCoins(context.Background(), sender.ID, sender.Username, req)

	assert.NoError(t, err, "failed to transfer coins")

	assert.NoError(t, db.First(&sender, sender.ID).Error, "failed to get sender")
	assert.NoError(t, db.First(&receiver, receiver.ID).Error, "failed to get receiver")

	assert.Equal(t, 900, sender.Coins, "unexpected sender coins count")
	assert.Equal(t, 1100, receiver.Coins, "unexpected receiver coins count")

	var transfer database.Transaction

	assert.NoError(t, db.Where("from_user_id = ? AND to_user_id = ?", sender.ID, receiver.ID).First(&transfer).Error, "failed to get transfer")

	assert.Equal(t, sender.ID, transfer.FromUserID, "unexpected sender id")
	assert.Equal(t, receiver.ID, transfer.ToUserID, "unexpected receiver id")

	assert.NoError(t, err)
}

func TestTransferCoins_NoReceiver(t *testing.T) {
	srv, db := getMockTransferService(t)

	sender := database.User{
		Username:     "test",
		PasswordHash: "test",
	}

	assert.NoError(t, db.Create(&sender).Error, "failed to create user")

	req := models.SendCoinRequest{
		ToUser: "receiver",
		Amount: 100,
	}

	err := srv.SendCoins(context.Background(), sender.ID, sender.Username, req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, services.ErrRecieverNotFound, "unexpected error")

	assert.NoError(t, db.First(&sender, sender.ID).Error, "failed to get sender")
	assert.Equal(t, 1000, sender.Coins, "unexpected sender coins count")

	var transfer database.Transaction

	assert.Error(t, db.Where("from_user_id = ? AND to_user_id = ?", sender.ID, 0).First(&transfer).Error, "transfer should not be created")
}

func TestTransferCoins_InsufficientFunds(t *testing.T) {
	srv, db := getMockTransferService(t)

	sender := database.User{
		Username:     "test1",
		PasswordHash: "test1",
	}
	receiver := database.User{
		Username:     "test2",
		PasswordHash: "test2",
	}

	assert.NoError(t, db.Create(&sender).Error, "failed to create user1")
	assert.NoError(t, db.Create(&receiver).Error, "failed to create user2")

	req := models.SendCoinRequest{
		ToUser: receiver.Username,
		Amount: 10000,
	}

	err := srv.SendCoins(context.Background(), sender.ID, sender.Username, req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInsufficientFunds, "unexpected error")
}

func TestTransferCoins_EmptyUsername(t *testing.T) {
	srv, _ := getMockTransferService(t)

	req := models.SendCoinRequest{
		ToUser: "",
		Amount: 100,
	}

	err := srv.SendCoins(context.Background(), 0, "", req)

	assert.Error(t, err)

	assert.ErrorIs(t, err, services.ErrToUserRequired, "unexpected error")
}

func TestTransferCoins_NegativeAmount(t *testing.T) {
	srv, _ := getMockTransferService(t)

	req := models.SendCoinRequest{
		ToUser: "receiver",
		Amount: -100,
	}

	err := srv.SendCoins(context.Background(), 0, "", req)

	assert.Error(t, err)

	assert.ErrorIs(t, err, services.ErrAmountBelowZero, "unexpected error")
}
