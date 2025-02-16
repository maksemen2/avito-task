package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maksemen2/avito-shop/internal/middleware"
	"github.com/maksemen2/avito-shop/internal/models"
	"github.com/maksemen2/avito-shop/internal/services"
)

func (h *RequestsHandler) SendCoin(c *gin.Context) {
	var req models.SendCoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrBadRequest))
		return
	}

	userID, _ := middleware.GetUserID(c)
	username, _ := middleware.GetUsername(c)

	err := h.transferService.SendCoins(c.Request.Context(), userID, username, req)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInternal):
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternal))
		default:
			// Могут возвращаться только ошибки с кодом ответа 400, поэтому можем себе позволить поступить так
			c.AbortWithStatusJSON(http.StatusBadRequest, models.NewDetailedErrorResponse(models.ErrBadRequest, err.Error()))
		}

		return
	}

	c.Status(http.StatusOK)
}
