package repository

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BaseRepository содержит общие зависимости и утилиты для всех репозиториев.
type BaseRepository struct {
	db     *gorm.DB
	Logger *zap.Logger
}

// WithTransaction выполняет переданную функцию в контексте транзакции.
func (r *BaseRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.DB(ctx).Transaction(fn)
}

// WrapError оборачивает ошибку с добавлением контекста.
func WrapError(context string, err error) error {
	return fmt.Errorf("%s: %w", context, err)
}

func (r *BaseRepository) DB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}
