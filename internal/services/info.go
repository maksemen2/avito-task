package services

import (
	"context"

	"github.com/maksemen2/avito-shop/internal/models"
	"github.com/maksemen2/avito-shop/internal/repository"
	"go.uber.org/zap"
)

type InfoService interface {
	GetInfo(ctx context.Context, userID uint) (models.InfoResponse, error)
}

type infoServiceImpl struct {
	repository repository.HolderRepository
	logger     *zap.Logger
}

func NewInfoService(repository repository.HolderRepository, logger *zap.Logger) InfoService {
	return &infoServiceImpl{repository: repository, logger: logger}
}

func (s *infoServiceImpl) GetInfo(ctx context.Context, userID uint) (models.InfoResponse, error) {
	balance, err := s.repository.User().GetBalance(ctx, userID)

	if err != nil {
		s.logger.Error("user not exists but token is valid", zap.Error(err), zap.Uint("userID", userID))
		return models.InfoResponse{}, ErrInternal
	}

	inventory, err := s.repository.Purchase().GetInventoryByUserID(ctx, userID)

	if err != nil {
		return models.InfoResponse{}, ErrInternal
	}

	coinHistory, err := s.repository.Transaction().GetHistoryByUserID(ctx, userID)

	if err != nil {
		return models.InfoResponse{}, ErrInternal
	}

	return models.InfoResponse{
		Coins:       balance,
		Inventory:   inventory,
		CoinHistory: coinHistory,
	}, nil
}
