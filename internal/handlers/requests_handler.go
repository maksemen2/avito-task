package handlers

import (
	"github.com/maksemen2/avito-shop/internal/repository"
	"github.com/maksemen2/avito-shop/internal/services"
	"github.com/maksemen2/avito-shop/pkg/auth"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RequestsHandler struct {
	JWTManager      *auth.JWTManager
	authService     services.AuthService
	transferService services.TransferService
	purchaseService services.PurchaseService
	infoService     services.InfoService
	logger          *zap.Logger
}

// NewRequestsHandler создаёт новый экземпляр RequestsHandler.
// Эта структура нужна для инъекции зависимостей в хендлеры.
func NewRequestsHandler(db *gorm.DB, jwtManager *auth.JWTManager, logger *zap.Logger) *RequestsHandler {
	repository := repository.NewHolderRepository(db, logger)

	return &RequestsHandler{
		JWTManager:      jwtManager,
		logger:          logger,
		authService:     services.NewAuthService(repository, jwtManager, logger),
		transferService: services.NewTransferService(repository, logger),
		purchaseService: services.NewPurchaseService(repository, logger),
		infoService:     services.NewInfoService(repository, logger),
	}
}
