package handlers

import (
	"github.com/maksemen2/avito-task/internal/auth"
	"github.com/maksemen2/avito-task/internal/dao"
	"gorm.io/gorm"
)

type RequestsHandler struct {
	dao        *dao.HolderDAO
	JWTManager *auth.JWTManager
}

func NewRequestsHandler(db *gorm.DB, jwtManager *auth.JWTManager) *RequestsHandler {
	return &RequestsHandler{dao: dao.NewHolderDAO(db), JWTManager: jwtManager}
}
