package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/maksemen2/avito-shop/config"
	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/maksemen2/avito-shop/internal/handlers"
	"github.com/maksemen2/avito-shop/internal/models"
	"github.com/maksemen2/avito-shop/internal/routes"
	"github.com/maksemen2/avito-shop/pkg/auth"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func setupTest(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	mockAuthConfig := config.AuthConfig{
		JwtKey:             "verySecretKey",
		TokenLifetimeHours: 72,
	}

	jwtManager := auth.NewJWTManager(mockAuthConfig)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormLogger.Default.LogMode(gormLogger.Silent)})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err = db.AutoMigrate(&database.User{}, &database.Purchase{}, &database.Transaction{}, &database.Good{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

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
		db.Create(&database.Good{Type: k, Price: v})
	}

	logger := zap.NewNop()

	reqHandler := handlers.NewRequestsHandler(db, jwtManager, logger)
	router := routes.SetupRoutes(reqHandler, logger, config.CorsConfig{AllowedOrigins: "*", AllowedMethods: "*", AllowedHeaders: "*", AllowCredientals: "true", MaxAge: "86300"})

	return router
}

func registerUser(t *testing.T, router *gin.Engine, username string) string {
	payload := fmt.Sprintf(`{"username": "%s", "password": "verySecurePassword"}`, username)
	req := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	var resp models.AuthResponse
	err := json.NewDecoder(recorder.Body).Decode(&resp)
	assert.NoError(t, err, "failed decoding auth response")

	return resp.Token
}

func getInfo(t *testing.T, router *gin.Engine, token string) models.InfoResponse {
	req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	var info models.InfoResponse
	err := json.NewDecoder(recorder.Body).Decode(&info)
	assert.NoError(t, err, "failed decoding info response")

	return info
}

func buyItem(router *gin.Engine, itemType string, token string) (int, models.ErrorResponse) {
	req := httptest.NewRequest(http.MethodGet, "/api/buy/"+itemType, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	var errResp models.ErrorResponse

	if recorder.Code != http.StatusOK {
		err := json.NewDecoder(recorder.Body).Decode(&errResp)
		if err != nil {
			return recorder.Code, models.ErrorResponse{}
		}
	}

	return recorder.Code, errResp
}

func transferCoins(router *gin.Engine, receiverUsername string, token string) (int, models.ErrorResponse) {
	payload := fmt.Sprintf(`{"toUser": "%s", "amount": 100}`, receiverUsername)
	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	var errResp models.ErrorResponse

	if recorder.Code != http.StatusOK {
		err := json.NewDecoder(recorder.Body).Decode(&errResp)
		if err != nil {
			return recorder.Code, models.ErrorResponse{}
		}
	}

	return recorder.Code, errResp
}

func TestE2EBuyMerch(t *testing.T) {
	router := setupTest(t)
	token := registerUser(t, router, "testUser")

	// Покупаем футболку
	code, _ := buyItem(router, "t-shirt", token)
	assert.Equal(t, http.StatusOK, code, "expected OK response for t-shirt purchase")

	// Получаем информацию о пользователе
	info := getInfo(t, router, token)
	assert.Len(t, info.Inventory, 1, "expected one purchase entry")
	assert.Equal(t, 920, info.Coins, "expected coin balance to be 920 after purchase")
}

func TestE2EBuyMerchWithNoCoins(t *testing.T) {
	router := setupTest(t)
	token := registerUser(t, router, "testUser")

	// Покупаем футболку 12 раз
	for i := 0; i < 12; i++ {
		code, _ := buyItem(router, "t-shirt", token)

		assert.Equal(t, http.StatusOK, code, "expected OK response for purchase #%d", i+1)
	}

	// 13-я покупка должна завершиться ошибкой (недостаточно монет)
	code, errResp := buyItem(router, "t-shirt", token)
	assert.Equal(t, http.StatusBadRequest, code, "expected BadRequest due to insufficient coins")

	assert.Contains(t, errResp.Errors, "insufficient", "expected error about insufficient coins")

	// Проверяем инвентарь пользователя
	info := getInfo(t, router, token)
	assert.Len(t, info.Inventory, 1, "expected one type of item in inventory")
	item := info.Inventory[0]
	assert.Equal(t, "t-shirt", item.Type, "expected item type t-shirt")
	assert.Equal(t, 12, item.Quantity, "expected quantity to be 12")
}

func TestTransferCoins(t *testing.T) {
	router := setupTest(t)

	// Создаем пользователей
	bankToken := registerUser(t, router, "bank")
	receiverOneToken := registerUser(t, router, "receiverOne")
	receiverTwoToken := registerUser(t, router, "receiverTwo")

	// Банк переводит 100 монет каждому из получателей
	for _, recipient := range []string{"receiverOne", "receiverTwo"} {
		code, _ := transferCoins(router, recipient, bankToken)
		assert.Equal(t, http.StatusOK, code, "expected OK response when sending coins to %s", recipient)
	}

	// Проверяем баланс банка
	info := getInfo(t, router, bankToken)
	assert.Equal(t, 800, info.Coins, "expected bank coin balance to be 800 after transfers")
	assert.Len(t, info.CoinHistory.Sent, 2, "expected two sent entries in coin history")

	for _, sent := range info.CoinHistory.Sent {
		assert.Equal(t, 100, sent.Amount, "expected transfer amount to be 100")
		assert.Contains(t, []string{"receiverOne", "receiverTwo"}, sent.ToUser, "unexpected recipient in coin history")
	}

	// Проверяем баланс получателей
	for _, token := range []string{receiverOneToken, receiverTwoToken} {
		info := getInfo(t, router, token)
		assert.Equal(t, 1100, info.Coins, "expected recipient coin balance to be 1100 after receiving coins")
		assert.Len(t, info.CoinHistory.Received, 1, "expected one received entry in coin history")
		assert.Equal(t, "bank", info.CoinHistory.Received[0].FromUser, "expected sender in received history to be bank")
	}

	// Первый получатель переводит 100 монет обратно банку
	code, _ := transferCoins(router, "bank", receiverOneToken)
	assert.Equal(t, http.StatusOK, code, "expected OK response for transfer back to bank")

	// Проверяем баланс банка после получения монет
	info = getInfo(t, router, bankToken)
	assert.Equal(t, 900, info.Coins, "expected bank coin balance to be 900 after receiving coins")
	assert.Len(t, info.CoinHistory.Sent, 2, "expected two sent entries to remain")
	assert.Len(t, info.CoinHistory.Received, 1, "expected one received entry in bank coin history")
	assert.Equal(t, "receiverOne", info.CoinHistory.Received[0].FromUser, "expected sender of returned coins to be receiverOne")
}
