// language: go
package services_test

import (
	"context"
	"testing"

	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/maksemen2/avito-shop/internal/repository"
	"github.com/maksemen2/avito-shop/internal/services"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func getMockInfoService(t *testing.T) (services.InfoService, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormLogger.Default.LogMode(gormLogger.Silent)})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	db.AutoMigrate(&database.User{}, &database.Purchase{}, &database.Transaction{}, &database.Good{})

	logger := zap.NewNop()
	holderRepo := repository.NewHolderRepository(db, logger)

	return services.NewInfoService(holderRepo, logger), db
}

func TestGetInfo_Success(t *testing.T) {
	infoService, db := getMockInfoService(t)

	user := database.User{
		Username:     "test",
		PasswordHash: "pass",
	}

	assert.NoError(t, db.Create(&user).Error, "failed to create user")

	resp, err := infoService.GetInfo(context.Background(), user.ID)

	assert.NoError(t, err)
	assert.Equal(t, 1000, resp.Coins, "unexpected coins count")
	assert.Len(t, resp.Inventory, 0, "items count should be 0")
	assert.Len(t, resp.CoinHistory.Received, 0, "received coins items count should be 0")
	assert.Len(t, resp.CoinHistory.Sent, 0, "sent coins items count should be 0")
}
