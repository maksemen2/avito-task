package dao

import (
	"github.com/maksemen2/avito-shop/internal/database"
	"gorm.io/gorm"
)

type UserDAO struct {
	DB *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{DB: db}
}

func (dao *UserDAO) Create(username string, passwordHash string) (*database.User, error) {
	user := &database.User{
		Username:     username,
		PasswordHash: passwordHash,
	}
	if result := dao.DB.Create(user); result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (dao *UserDAO) GetByID(id uint) (*database.User, error) {
	var user database.User
	if result := dao.DB.First(&user, id); result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (dao *UserDAO) GetByUsername(username string) (*database.User, error) {
	var user database.User
	if result := dao.DB.Where("username = ?", username).First(&user); result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (dao *UserDAO) IsExists(username string) (bool, error) {
	var count int64
	if err := dao.DB.Model(&database.User{}).
		Where("username = ?", username).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (dao *UserDAO) GetBalance(id uint) (int, error) {
	var balance int
	if result := dao.DB.Model(&database.User{}).Select("coins").Where("id = ?", id).Scan(&balance); result.Error != nil {
		return 0, result.Error
	}
	return balance, nil
}
