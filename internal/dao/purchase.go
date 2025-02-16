package dao

import (
	"fmt"

	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/maksemen2/avito-shop/internal/models"
	"gorm.io/gorm"
)

type PurchaseDAO struct {
	DB *gorm.DB
}

func NewPurchaseDAO(db *gorm.DB) *PurchaseDAO {
	return &PurchaseDAO{DB: db}
}

// GetInventoryByUserID возвращает инвентарь пользователя в виде json модели Item по его айди.
func (dao *PurchaseDAO) GetInventoryByUserID(userID uint) ([]models.Item, error) {
	var items []models.Item

	err := dao.DB.
		Model(&database.Purchase{}).
		Select("goods.type as type, COUNT(purchases.id) as quantity").
		Joins("LEFT JOIN goods ON goods.id = purchases.good_id").
		Where("purchases.user_id = ?", userID).
		Group("goods.type").
		Order("goods.type ASC").
		Scan(&items).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	if items == nil {
		return []models.Item{}, nil
	}

	return items, nil
}
