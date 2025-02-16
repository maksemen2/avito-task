package repository

import (
	"context"

	"github.com/maksemen2/avito-shop/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TransactionRepository описывает операции для получения истории транзакций.
type TransactionRepository interface {
	GetHistoryByUserID(ctx context.Context, userID uint) (models.CoinHistory, error)
}

// GormTransactionRepository реализует TransactionRepository.
type GormTransactionRepository struct {
	BaseRepository
}

func NewTransactionRepository(db *gorm.DB, logger *zap.Logger) TransactionRepository {
	return &GormTransactionRepository{
		BaseRepository: BaseRepository{
			db:     db,
			Logger: logger,
		},
	}
}

func (r *GormTransactionRepository) GetHistoryByUserID(ctx context.Context, userID uint) (models.CoinHistory, error) {
	var received []models.ReceivedCoins

	var sent []models.SentCoins

	if err := r.DB(ctx).Table("transactions").
		Select("users.username as from_user, transactions.amount, transactions.created_at").
		Joins("JOIN users ON transactions.from_user_id = users.id").
		Where("transactions.to_user_id = ?", userID).
		Order("transactions.created_at DESC").
		Scan(&received).Error; err != nil {
		r.Logger.Error("failed to get received coins", zap.Uint("userID", userID), zap.Error(err))
		return models.CoinHistory{}, WrapError(ErrGetHistory.Error(), err)
	}

	if err := r.DB(ctx).Table("transactions").
		Select("users.username as to_user, transactions.amount, transactions.created_at").
		Joins("JOIN users ON transactions.to_user_id = users.id").
		Where("transactions.from_user_id = ?", userID).
		Order("transactions.created_at DESC").
		Scan(&sent).Error; err != nil {
		r.Logger.Error("failed to get sent coins", zap.Uint("userID", userID), zap.Error(err))
		return models.CoinHistory{}, WrapError(ErrGetHistory.Error(), err)
	}

	if received == nil {
		received = make([]models.ReceivedCoins, 0)
	}

	if sent == nil {
		sent = make([]models.SentCoins, 0)
	}

	return models.CoinHistory{
		Received: received,
		Sent:     sent,
	}, nil
}
