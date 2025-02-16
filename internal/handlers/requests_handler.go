package handlers

import (
	"github.com/maksemen2/avito-shop/internal/auth"
	"github.com/maksemen2/avito-shop/internal/dao"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RequestsHandler struct {
	dao        *dao.HolderDAO
	JWTManager *auth.JWTManager
	logger     *zap.Logger
}

// NewRequestsHandler создаёт новый экземпляр RequestsHandler.
// Эта структура нужна для инъекции зависимостей в хендлеры.
func NewRequestsHandler(db *gorm.DB, jwtManager *auth.JWTManager, logger *zap.Logger) *RequestsHandler {
	return &RequestsHandler{dao: dao.NewHolderDAO(db), JWTManager: jwtManager, logger: logger}
}
