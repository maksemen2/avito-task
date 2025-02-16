package repository

import (
	"context"

	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/maksemen2/avito-shop/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PurchaseRepository описывает операции для покупок.
type PurchaseRepository interface {
	GetInventoryByUserID(ctx context.Context, userID uint) ([]models.Item, error)
}

// GormPurchaseRepository реализует PurchaseRepository.
type GormPurchaseRepository struct {
	BaseRepository
}

func NewPurchaseRepository(db *gorm.DB, logger *zap.Logger) PurchaseRepository {
	return &GormPurchaseRepository{
		BaseRepository: BaseRepository{
			db:     db,
			Logger: logger,
		},
	}
}

func (r *GormPurchaseRepository) GetInventoryByUserID(ctx context.Context, userID uint) ([]models.Item, error) {
	var items []models.Item
	if err := r.DB(ctx).
		Model(&database.Purchase{}).
		Select("goods.type as type, COUNT(purchases.id) as quantity").
		Joins("LEFT JOIN goods ON goods.id = purchases.good_id").
		Where("purchases.user_id = ?", userID).
		Group("goods.type").
		Order("goods.type ASC").
		Scan(&items).Error; err != nil {
		r.Logger.Error("failed to get inventory", zap.Uint("userID", userID), zap.Error(err))
		return nil, WrapError(ErrGetInventory.Error(), err)
	}

	if items == nil {
		return []models.Item{}, nil
	}

	return items, nil
}
