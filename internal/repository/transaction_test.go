package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/maksemen2/avito-shop/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func setupTestTransactionRepository(t *testing.T) (repository.TransactionRepository, *gorm.DB) {
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
	txRepo := repository.NewTransactionRepository(db, logger)

	return txRepo, db
}

func TestGetHistoryByUserID_Success(t *testing.T) {
	txRepo, db := setupTestTransactionRepository(t)
	ctx := context.Background()

	alice := database.User{
		Username: "alice",
		Coins:    100,
	}
	bob := database.User{
		Username: "bob",
		Coins:    50,
	}

	assert.NoError(t, db.Create(&alice).Error, "failed to create alice")
	assert.NoError(t, db.Create(&bob).Error, "failed to create bob")

	tx1 := database.Transaction{
		FromUserID: bob.ID,
		ToUserID:   alice.ID,
		Amount:     50,
		CreatedAt:  time.Now().Add(-time.Minute),
	}
	tx2 := database.Transaction{
		FromUserID: alice.ID,
		ToUserID:   bob.ID,
		Amount:     30,
		CreatedAt:  time.Now(),
	}

	assert.NoError(t, db.Create(&tx1).Error, "failed to create transaction from bob to alice")
	assert.NoError(t, db.Create(&tx2).Error, "failed to create transaction from alice to bob")

	history, err := txRepo.GetHistoryByUserID(ctx, alice.ID)
	assert.NoError(t, err, "expected no error retrieving coin history")

	sent, received := history.Sent, history.Received
	assert.Equal(t, 1, len(sent), "expected one sent transaction")
	assert.Equal(t, 1, len(received), "expected one received transaction")

	recTx := received[0]
	assert.Equal(t, bob.Username, recTx.FromUser, "expected sender username to match")
	assert.Equal(t, tx1.Amount, recTx.Amount, "expected amount to match")

	sentTx := sent[0]
	assert.Equal(t, bob.Username, sentTx.ToUser, "expected receiver username to match")
	assert.Equal(t, tx2.Amount, sentTx.Amount, "expected amount to match")
}
