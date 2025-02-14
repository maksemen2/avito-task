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
		Select("item_name as type, COUNT(*) as quantity").
		Where("user_id = ?", userID).
		Group("item_name").
		Scan(&items); result.Error != nil {
		return nil, result.Error
	}
	if items == nil {
		items = []models.Item{}
	}
	return items, nil
}
