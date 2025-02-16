package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maksemen2/avito-shop/internal/middleware"
	"github.com/maksemen2/avito-shop/internal/models"
	"github.com/maksemen2/avito-shop/internal/services"
)

func (h *RequestsHandler) BuyItem(c *gin.Context) {
	item := c.Param("item")
	userID, _ := middleware.GetUserID(c)
	err := h.purchaseService.BuyGood(c.Request.Context(), userID, item)

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
