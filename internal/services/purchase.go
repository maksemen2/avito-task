package services

import (
	"context"
	"errors"

	"github.com/maksemen2/avito-shop/internal/repository"
	"go.uber.org/zap"
)

type PurchaseService interface {
	BuyGood(ctx context.Context, userID uint, goodName string) error
}

type purchaseServiceImpl struct {
	repository repository.HolderRepository
	logger     *zap.Logger
}

func NewPurchaseService(repository repository.HolderRepository, logger *zap.Logger) PurchaseService {
	return &purchaseServiceImpl{repository: repository, logger: logger}
}

func (s *purchaseServiceImpl) BuyGood(ctx context.Context, userID uint, itemType string) error {
	if itemType == "" {
		return ErrItemTypeRequired
	}

	good, err := s.repository.Good().GetByName(ctx, itemType)

	if err != nil {
		if errors.Is(err, repository.ErrGoodNotFound) {
			return ErrItemNotFound
		} else {
			return ErrInternal
		}
	}

	if err := s.repository.BuyItem(ctx, userID, good.ID, good.Price); err != nil {
		if errors.Is(err, repository.ErrInsufficientFunds) {
			return ErrInsufficientFunds
		} else {
			return ErrInternal
		}
	}

	return nil
}
