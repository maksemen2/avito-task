package services

import (
	"context"
	"errors"

	"github.com/maksemen2/avito-shop/internal/models"
	"github.com/maksemen2/avito-shop/internal/repository"
	"go.uber.org/zap"
)

type TransferService interface {
	SendCoins(ctx context.Context, senderID uint, senderUsername string, req models.SendCoinRequest) error
}

type transferServiceImpl struct {
	repository repository.HolderRepository
	logger     *zap.Logger
}

func NewTransferService(repository repository.HolderRepository, logger *zap.Logger) TransferService {
	return &transferServiceImpl{repository: repository, logger: logger}
}

func (s *transferServiceImpl) SendCoins(ctx context.Context, senderID uint, senderUsername string, req models.SendCoinRequest) error {
	if req.ToUser == "" {
		return ErrToUserRequired
	}

	if req.Amount <= 0 {
		return ErrAmountBelowZero
	}

	if senderUsername == req.ToUser {
		return ErrCantSelfTransfer
	}

	receiverID, err := s.repository.User().GetIDByUsername(ctx, req.ToUser)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrRecieverNotFound
		}

		return ErrInternal
	}

	err = s.repository.TransferCoins(ctx, senderID, receiverID, req.Amount)

	if err != nil {
		if errors.Is(err, repository.ErrInsufficientFunds) {
			return ErrInsufficientFunds
		} else {
			return ErrInternal
		}
	}

	return nil
}
