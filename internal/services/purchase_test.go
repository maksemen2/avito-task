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

func getMockPurchaseService(t *testing.T) (services.PurchaseService, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormLogger.Default.LogMode(gormLogger.Silent)})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	db.AutoMigrate(&database.User{}, &database.Purchase{}, &database.Transaction{}, &database.Good{})

	goods := map[string]int{
		"t-shirt":    80,
		"cup":        20,
		"book":       50,
		"pen":        20,
		"powerbank":  200,
		"hoody":      300,
		"umbrella":   200,
		"socks":      10,
		"wallet":     50,
		"pink-hoody": 500,
	}

	for k, v := range goods {
		err := db.Create(&database.Good{Type: k, Price: v}).Error
		if err != nil {
			t.Fatalf("failed to create good: %v", err)
		}
	}

	logger := zap.NewNop()
	holderRepo := repository.NewHolderRepository(db, logger)

	return services.NewPurchaseService(holderRepo, logger), db
}

func TestPurchaseItem_Success(t *testing.T) {
	srv, db := getMockPurchaseService(t)

	user := database.User{
		Username:     "test",
		PasswordHash: "test",
	}

	assert.NoError(t, db.Create(&user).Error, "failed to create user")

	err := srv.BuyGood(context.Background(), user.ID, "t-shirt")

	assert.NoError(t, err)

	var updatedUser database.User

	assert.NoError(t, db.First(&updatedUser, user.ID).Error, "failed to fetch user")
	assert.Equal(t, 920, updatedUser.Coins, "user's coins should be deducted")

	var purchase database.Purchase

	var tShirt database.Good

	assert.NoError(t, db.Where("type = ?", "t-shirt").First(&tShirt).Error, "failed to fetch good")

	assert.NoError(t, db.Where("user_id = ?", user.ID).First(&purchase).Error, "failed to fetch purchase")
	assert.Equal(t, tShirt.ID, purchase.GoodID, "unexpected good bought")
}

func TestPurchaseItem_NoCoins(t *testing.T) {
	srv, db := getMockPurchaseService(t)
	ctx := context.Background()
	user := database.User{
		Username:     "test",
		PasswordHash: "test",
	}

	assert.NoError(t, db.Create(&user).Error, "failed to create user")

	for i := 0; i < 12; i++ {
		assert.NoError(t, srv.BuyGood(ctx, user.ID, "t-shirt"), "failed to buy tshirt")
	}

	err := srv.BuyGood(context.Background(), user.ID, "t-shirt")

	assert.Error(t, err)
	assert.Equal(t, services.ErrInsufficientFunds, err)
}

func TestPurchaseItem_NoGood(t *testing.T) {
	srv, db := getMockPurchaseService(t)

	user := database.User{
		Username:     "test",
		PasswordHash: "test",
		Coins:        1000,
	}

	assert.NoError(t, db.Create(&user).Error, "failed to create user")

	err := srv.BuyGood(context.Background(), user.ID, "non-existing-good")

	assert.Error(t, err)
	assert.Equal(t, services.ErrItemNotFound, err)
}
