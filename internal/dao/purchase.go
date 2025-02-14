package dao

import (
	"github.com/maksemen2/avito-task/internal/database"
	"github.com/maksemen2/avito-task/internal/models"
	"gorm.io/gorm"
)

type PurchaseDAO struct {
	DB *gorm.DB
}

func NewPurchaseDAO(db *gorm.DB) *PurchaseDAO {
	return &PurchaseDAO{DB: db}
}

func (dao *PurchaseDAO) GetInventoryByUserID(userID uint) ([]models.Item, error) {
	var items []models.Item
	if result := dao.DB.Model(&database.Purchase{}).
		Select("type, SUM(quantity) as quantity").
		Where("user_id = ?", userID).
		Group("type").
		Scan(&items); result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}
