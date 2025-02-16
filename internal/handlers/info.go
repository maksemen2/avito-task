package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maksemen2/avito-shop/internal/middleware"
	"github.com/maksemen2/avito-shop/internal/models"
)

func (h *RequestsHandler) GetInfo(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	resp, err := h.infoService.GetInfo(c.Request.Context(), userID)

	if err != nil {
		// может быть только services.ErrInternal
		c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternal))
	}

	c.JSON(http.StatusOK, resp)
}
