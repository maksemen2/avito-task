package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maksemen2/avito-shop/internal/models"
	"github.com/maksemen2/avito-shop/internal/services"
)

func (h *RequestsHandler) Authenticate(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, models.NewDetailedErrorResponse(models.ErrBadRequest, "invalid request"))
		return
	}

	resp, err := h.authService.Authenticate(c.Request.Context(), req)
	if err != nil {
		errDetail := err.Error()

		switch {
		case errors.Is(err, services.ErrUserPassRequired):
			c.AbortWithStatusJSON(http.StatusBadRequest, models.NewDetailedErrorResponse(models.ErrBadRequest, errDetail))
		case errors.Is(err, services.ErrAuthFailed):
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.NewDetailedErrorResponse(models.ErrUnauthorized, errDetail))
		case errors.Is(err, services.ErrInternal):
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternal))
		}

		return
	}

	c.JSON(http.StatusOK, resp)
}
