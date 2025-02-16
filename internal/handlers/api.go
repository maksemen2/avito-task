package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maksemen2/avito-shop/internal/dao"
	"github.com/maksemen2/avito-shop/internal/middleware"
	"github.com/maksemen2/avito-shop/internal/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func (h *RequestsHandler) Authenticate(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.NewDetailedErrorResponse(models.ErrBadRequest, "invalid request"))
		return
	}

	user, err := h.dao.User.GetByUsername(req.Username)
	wasCreated := false

	if err != nil {
		if errors.Is(err, dao.ErrUserNotFound) {
			// Если пользователя нет в бд - регистрируем его
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				h.logger.Error("Password hash generation failed", zap.Error(err))
				c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternal))

				return
			}

			if user, err = h.dao.User.Create(req.Username, string(passwordHash)); err != nil {
				h.logger.Error("User creation failed",
					zap.String("username", req.Username),
					zap.Error(err))
				c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternal))

				return
			}

			wasCreated = true
		} else {
			h.logger.Error("DB error in auth",
				zap.String("username", req.Username),
				zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternal))

			return
		}
	}

	// Сравнение хешей - довольно дорогая операция, поэтому производим её только если пользователь не был зарегестрирован
	// в процессе обработки этого запроса, так можно немного повысить производительность
	if !wasCreated {
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewDetailedErrorResponse(models.ErrUnauthorized, "authentication failed"))
			return
		}
	}

	token, err := h.JWTManager.GenerateToken(user.ID, user.Username)
	if err != nil {
		h.logger.Error("Token generation failed",
			zap.Uint("userID", user.ID),
			zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternal))

		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{Token: token})
}

func (h *RequestsHandler) SendCoin(c *gin.Context) {
	var req models.SendCoinRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ToUser == "" || req.Amount <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.NewDetailedErrorResponse(models.ErrBadRequest, "invalid request"))
		return
	}

	userID, _ := middleware.GetUserID(c)

	receiverID, err := h.dao.User.GetIDByUsername(req.ToUser)

	if err != nil {
		if errors.Is(err, dao.ErrUserNotFound) {
			c.AbortWithStatusJSON(http.StatusBadRequest, models.NewDetailedErrorResponse(models.ErrBadRequest, "recipient not found"))
		} else {
			h.logger.Error("User lookup failed",
				zap.String("username", req.ToUser),
				zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternal))
		}

		return
	}

	if receiverID == userID { // Переводить самому себе нельзя
		c.AbortWithStatusJSON(http.StatusBadRequest, models.NewDetailedErrorResponse(models.ErrBadRequest, "self transfer not allowed"))
		return
	}

	if err := h.dao.TransferCoins(userID, receiverID, req.Amount); err != nil {
		if errors.Is(err, dao.ErrInsufficientFunds) {
			c.AbortWithStatusJSON(http.StatusBadRequest, models.NewDetailedErrorResponse(models.ErrBadRequest, "insufficient funds"))
		} else if errors.Is(err, dao.ErrUserNotFound) {
			c.AbortWithStatusJSON(http.StatusBadRequest, models.NewDetailedErrorResponse(models.ErrBadRequest, "recipient not found"))
		} else {
			h.logger.Error("Transfer failed",
				zap.Uint("userID", userID),
				zap.String("toUser", req.ToUser),
				zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternal))
		}

		return
	}

	c.Status(http.StatusOK)
}

func (h *RequestsHandler) BuyItem(c *gin.Context) {
	itemName := c.Param("item")
	if itemName == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.NewDetailedErrorResponse(models.ErrBadRequest, "item required"))
		return
	}

	userID, _ := middleware.GetUserID(c)

	good, err := h.dao.Good.GetByName(itemName) // сразу проверяем существование товара и заодно получаем его айди с цену

	if err != nil {
		if errors.Is(err, dao.ErrGoodNotFound) {
			h.logger.Warn("Item not found", zap.String("item", itemName))
			c.AbortWithStatusJSON(http.StatusBadRequest, models.NewDetailedErrorResponse(models.ErrBadRequest, "item not found"))
		} else {
			h.logger.Error("Item lookup failed",
				zap.String("item", itemName),
				zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternal))
		}

		return
	}

	if err := h.dao.BuyItem(userID, good.ID, good.Price); err != nil {
		if errors.Is(err, dao.ErrInsufficientFunds) {
			c.AbortWithStatusJSON(http.StatusBadRequest, models.NewDetailedErrorResponse(models.ErrBadRequest, "insufficient funds"))
		} else {
			h.logger.Error("Purchase failed",
				zap.Uint("userID", userID),
				zap.String("item", itemName),
				zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternal))
		}

		return
	}

	c.Status(http.StatusOK)
}

func (h *RequestsHandler) GetInfo(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	balance, err := h.dao.User.GetBalance(userID)
	if err != nil {
		if errors.Is(err, dao.ErrUserNotFound) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewDetailedErrorResponse(models.ErrUnauthorized, "user not found"))
		} else {
			h.logger.Error("Balance check failed",
				zap.Uint("userID", userID),
				zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternal))
		}

		return
	}

	inventory, err := h.dao.Purchase.GetInventoryByUserID(userID)
	if err != nil {
		h.logger.Error("Inventory check failed",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternal))

		return
	}

	receivedCoins, sentCoins, err := h.dao.Transaction.GetHistoryByUserID(userID)
	if err != nil {
		h.logger.Error("History check failed",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternal))

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
