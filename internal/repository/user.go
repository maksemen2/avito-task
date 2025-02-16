package repository

import (
	"context"
	"errors"

	"github.com/maksemen2/avito-shop/internal/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserRepository описывает CRUD операции для пользователей.
type UserRepository interface {
	Create(ctx context.Context, username, passwordHash string) (*database.User, error)
	GetByID(ctx context.Context, id uint) (*database.User, error)
	GetByUsername(ctx context.Context, username string) (*database.User, error)
	GetBalance(ctx context.Context, id uint) (int, error)
	GetIDByUsername(ctx context.Context, username string) (uint, error)
}

// GormUserRepository – реализация UserRepository для GORM.
type GormUserRepository struct {
	BaseRepository
}

func NewUserRepository(db *gorm.DB, logger *zap.Logger) UserRepository {
	return &GormUserRepository{
		BaseRepository: BaseRepository{
			db:     db,
			Logger: logger,
		},
	}
}

func (r *GormUserRepository) Create(ctx context.Context, username, passwordHash string) (*database.User, error) {
	user := &database.User{
		Username:     username,
		PasswordHash: passwordHash,
	}

	if err := r.DB(ctx).Create(user).Error; err != nil {
		r.Logger.Error("failed to create user", zap.String("username", username), zap.Error(err))
		return nil, WrapError(ErrCreateUser.Error(), err)
	}

	return user, nil
}

func (r *GormUserRepository) GetByID(ctx context.Context, id uint) (*database.User, error) {
	var user database.User
	if err := r.DB(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		r.Logger.Error("failed to get user", zap.Uint("userID", id), zap.Error(err))

		return nil, WrapError(ErrGetUser.Error(), err)
	}

	return &user, nil
}

func (r *GormUserRepository) GetByUsername(ctx context.Context, username string) (*database.User, error) {
	var user database.User
	if err := r.DB(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		r.Logger.Error("failed to get user", zap.String("username", username), zap.Error(err))

		return nil, WrapError(ErrGetUser.Error(), err)
	}

	return &user, nil
}

func (r *GormUserRepository) GetBalance(ctx context.Context, id uint) (int, error) {
	var balance int
	if err := r.DB(ctx).Model(&database.User{}).Select("coins").Where("id = ?", id).Scan(&balance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrUserNotFound
		}

		r.Logger.Error("failed to get balance", zap.Uint("userID", id), zap.Error(err))

		return 0, WrapError(ErrGetBalance.Error(), err)
	}

	return balance, nil
}

func (r *GormUserRepository) GetIDByUsername(ctx context.Context, username string) (uint, error) {
	// ...existing code...
	var user database.User
	if err := r.DB(ctx).
		Select("id").
		Where("username = ?", username).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrUserNotFound
		}

		r.Logger.Error("failed to get user ID", zap.String("username", username), zap.Error(err))

		return 0, WrapError(ErrGetUser.Error(), err)
	}

	return user.ID, nil
}
