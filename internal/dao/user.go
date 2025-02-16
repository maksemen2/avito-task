package dao

import (
	"errors"
	"fmt"

	"github.com/maksemen2/avito-shop/internal/database"
	"gorm.io/gorm"
)

type UserDAO struct {
	DB *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{DB: db}
}

// Create создает нового пользователя
func (dao *UserDAO) Create(username, passwordHash string) (*database.User, error) {
	user := &database.User{
		Username:     username,
		PasswordHash: passwordHash,
	}

	err := dao.DB.Create(user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetByID возвращает пользователя по ID с контролем контекста
func (dao *UserDAO) GetByID(id uint) (*database.User, error) {
	var user database.User

	err := dao.DB.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByUsername ищет пользователя по имени с использованием индекса
func (dao *UserDAO) GetByUsername(username string) (*database.User, error) {
	var user database.User
	err := dao.DB.
		Where("username = ?", username).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrUserNotFound, username)
		}

		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetBalance возвращает баланс с проверкой существования пользователя
func (dao *UserDAO) GetBalance(id uint) (int, error) {
	var balance int
	if result := dao.DB.Model(&database.User{}).Select("coins").Where("id = ?", id).Scan(&balance); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, ErrUserNotFound
		}

		return 0, result.Error
	}

	return balance, nil
}

// GetIDByUsername возвращает ID пользователя по его юзернейму
func (dao *UserDAO) GetIDByUsername(username string) (uint, error) {
	var id uint
	if result := dao.DB.Model(&database.User{}).Select("id").Where("username = ?", username).Scan(&id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, ErrUserNotFound
		}

		return 0, result.Error
	}

	return id, nil
}
