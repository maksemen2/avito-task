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
	"github.com/maksemen2/avito-shop/internal/auth"
	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/maksemen2/avito-shop/internal/handlers"
	"github.com/maksemen2/avito-shop/internal/models"
	"github.com/maksemen2/avito-shop/internal/routes"
	"github.com/maksemen2/avito-shop/pkg/logger"
	"github.com/stretchr/testify/assert"
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

	mockLoggerConfig := config.LoggerConfig{
		Level:    "fatal",
		FilePath: "",
	}

	logger := logger.MustLoad(mockLoggerConfig)

	reqHandler := handlers.NewRequestsHandler(db, jwtManager, logger)
	router := routes.SetupRoutes(reqHandler, logger, config.CorsConfig{})

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

func TestE2EBuyMerch(t *testing.T) {
	router := setupTest(t)
	token := registerUser(t, router, "testUser")

	// Покупаем футболку
	req := httptest.NewRequest(http.MethodGet, "/api/buy/t-shirt", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code, "expected OK response for t-shirt purchase")

	// Получаем информацию о пользователе
	req = httptest.NewRequest(http.MethodGet, "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code, "expected OK response for info request")

	var info models.InfoResponse
	err := json.NewDecoder(recorder.Body).Decode(&info)
	assert.NoError(t, err, "failed decoding info response")
	assert.Len(t, info.Inventory, 1, "expected one purchase entry")
	assert.Equal(t, 920, info.Coins, "expected coin balance to be 920 after purchase")
}

func TestE2EBuyMerchWithNoCoins(t *testing.T) {
	router := setupTest(t)
	token := registerUser(t, router, "testUser")

	// Покупаем футболку 12 раз
	for i := 0; i < 12; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/buy/t-shirt", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code, "expected OK response for purchase #%d", i+1)
	}

	// 13-я покупка должна завершиться ошибкой (недостаточно монет)
	req := httptest.NewRequest(http.MethodGet, "/api/buy/t-shirt", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code, "expected BadRequest due to insufficient coins")

	var errResp models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&errResp)
	assert.NoError(t, err, "failed decoding error response")
	assert.Contains(t, errResp.Errors, "insufficient", "expected error about insufficient coins")

	// Проверяем инвентарь пользователя
	req = httptest.NewRequest(http.MethodGet, "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code, "expected OK response for info request")

	var info models.InfoResponse
	err = json.NewDecoder(recorder.Body).Decode(&info)
	assert.NoError(t, err, "failed decoding info response")
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
		payload := fmt.Sprintf(`{"toUser": "%s", "amount": 100}`, recipient)
		req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+bankToken)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code, "expected OK response when sending coins to %s", recipient)
	}

	// Проверяем баланс банка
	req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+bankToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code, "expected OK response for bank info request")

	var info models.InfoResponse
	err := json.NewDecoder(recorder.Body).Decode(&info)
	assert.NoError(t, err, "failed decoding bank info response")
	assert.Equal(t, 800, info.Coins, "expected bank coin balance to be 800 after transfers")
	assert.Len(t, info.CoinHistory.Sent, 2, "expected two sent entries in coin history")

	for _, sent := range info.CoinHistory.Sent {
		assert.Equal(t, 100, sent.Amount, "expected transfer amount to be 100")
		assert.Contains(t, []string{"receiverOne", "receiverTwo"}, sent.ToUser, "unexpected recipient in coin history")
	}

	// Проверяем баланс получателей
	for _, token := range []string{receiverOneToken, receiverTwoToken} {
		req = httptest.NewRequest(http.MethodGet, "/api/info", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		recorder = httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code, "expected OK response for recipient info request")

		err = json.NewDecoder(recorder.Body).Decode(&info)
		assert.NoError(t, err, "failed decoding recipient info response")
		assert.Equal(t, 1100, info.Coins, "expected recipient coin balance to be 1100 after receiving coins")
		assert.Len(t, info.CoinHistory.Received, 1, "expected one received entry in coin history")
		assert.Equal(t, "bank", info.CoinHistory.Received[0].FromUser, "expected sender in received history to be bank")
	}

	// Первый получатель переводит 100 монет обратно банку
	payload := `{"toUser": "bank", "amount": 100}`
	req = httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+receiverOneToken)

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code, "expected OK response for transfer back to bank")

	// Проверяем баланс банка после получения монет
	req = httptest.NewRequest(http.MethodGet, "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+bankToken)

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code, "expected OK response for bank info request after receipt")

	err = json.NewDecoder(recorder.Body).Decode(&info)
	assert.NoError(t, err, "failed decoding bank info response after receipt")
	assert.Equal(t, 900, info.Coins, "expected bank coin balance to be 900 after receiving coins")
	assert.Len(t, info.CoinHistory.Sent, 2, "expected two sent entries to remain")
	assert.Len(t, info.CoinHistory.Received, 1, "expected one received entry in bank coin history")
	assert.Equal(t, "receiverOne", info.CoinHistory.Received[0].FromUser, "expected sender of returned coins to be receiverOne")
}
