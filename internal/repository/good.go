package repository

import (
	"context"
	"errors"

	"github.com/maksemen2/avito-shop/internal/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GoodRepository описывает операции для работы с товарами.
type GoodRepository interface {
	GetByName(ctx context.Context, name string) (*database.Good, error)
}

// GormGoodRepository реализует GoodRepository.
type GormGoodRepository struct {
	BaseRepository
}

func NewGoodRepository(db *gorm.DB, logger *zap.Logger) GoodRepository {
	return &GormGoodRepository{
		BaseRepository: BaseRepository{
			db:     db,
			Logger: logger,
		},
	}
}

func (r *GormGoodRepository) GetByName(ctx context.Context, name string) (*database.Good, error) {
	var good database.Good
	if err := r.DB(ctx).Where("type = ?", name).First(&good).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGoodNotFound
		}

		r.Logger.Error("failed to get good", zap.String("name", name), zap.Error(err))

		return nil, WrapError(ErrGetGood.Error(), err)
	}

	return &good, nil
}
