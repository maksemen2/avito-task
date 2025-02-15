package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maksemen2/avito-shop/internal/middleware"
	"github.com/maksemen2/avito-shop/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (h *RequestsHandler) Authenticate(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrBadRequest, "username and password are required"))

		return
	}

	user, err := h.dao.User.GetByUsername(req.Username)
	if err != nil {
		passwordBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Errors: models.ErrIternal})

			return
		}

		user, err = h.dao.User.Create(req.Username, string(passwordBytes))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Errors: models.ErrIternal})

			return
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized, "invalid password"))

		return
	}

	token, err := h.JWTManager.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Errors: models.ErrIternal})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{Token: token})
}

func (h *RequestsHandler) SendCoin(c *gin.Context) {
	var req models.SendCoinRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ToUser == "" || req.Amount <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrBadRequest, "to_user and amount are required"))
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Errors: models.ErrUnauthorized})
		return
	}

	user, err := h.dao.User.GetByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Errors: models.ErrUnauthorized})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Errors: models.ErrIternal})
		}

		return
	}

	if user.Coins < req.Amount {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrBadRequest, "not enough coins"))
		return
	}

	if user.Username == req.ToUser {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrBadRequest, "cannot send coins to yourself"))
		return
	}

	recipient, err := h.dao.User.GetByUsername(req.ToUser)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.AbortWithStatusJSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrBadRequest, "recipient not found"))
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Errors: models.ErrIternal})
		}

		return
	}

	if err := h.dao.TransferCoins(userID, recipient.ID, req.Amount); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Errors: models.ErrIternal})
		return
	}

	c.Status(http.StatusOK)
}

// nolint:gochecknoglobals
// Да, использование глобальных переменных - не лучшая практика, однако она эта мапа используется только в методе BuyItem, и в задании не указано, что названия, цены и ассортимент товаров могут изменяться
var goods = map[string]int{
	"t-shirt":    80,
	"cup":        20,
	"book":       50,
	"pen":        10,
	"powerbank":  200,
	"hoody":      300,
	"umbrella":   200,
	"socks":      10,
	"wallet":     50,
	"pink-hoody": 500,
}

func (h *RequestsHandler) BuyItem(c *gin.Context) {
	itemName := c.Param("item")
	if itemName == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrBadRequest, "item name is required"))
		return
	}

	price, found := goods[itemName]
	if !found {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrBadRequest, "item not found"))
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Errors: models.ErrUnauthorized})
		return
	}

	user, err := h.dao.User.GetByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Errors: models.ErrUnauthorized})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Errors: models.ErrIternal})
		}

		return
	}

	if user.Coins < price {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrBadRequest, "not enough coins"))
		return
	}

	if err := h.dao.BuyItem(userID, itemName); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Errors: models.ErrIternal})
		return
	}

	c.Status(http.StatusOK)
}

func (h *RequestsHandler) GetInfo(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Errors: models.ErrUnauthorized})
		return
	}

	balance, err := h.dao.User.GetBalance(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{Errors: models.ErrUnauthorized})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Errors: models.ErrIternal})
		}

		return
	}

	inventory, err := h.dao.Purchase.GetInventoryByUserID(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Errors: models.ErrIternal})
		return
	}

	receivedCoins, sentCoins, err := h.dao.Transaction.GetHistoryByUserID(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{Errors: models.ErrIternal})
		return
	}

	c.JSON(http.StatusOK, models.InfoResponse{
		Coins:     balance,
		Inventory: inventory,
		CoinHistory: models.CoinHistory{
			Sent:     sentCoins,
			Received: receivedCoins,
		},
	})
}
