package dao

import (
	"github.com/maksemen2/avito-shop/internal/database"
	"gorm.io/gorm"
)

type GoodDAO struct {
	DB *gorm.DB
}

func NewGoodDAO(db *gorm.DB) *GoodDAO {
	return &GoodDAO{DB: db}
}

// GetByName возвращает товар по его названию.
func (dao *GoodDAO) GetByName(name string) (*database.Good, error) {
	var good database.Good
	if result := dao.DB.Where("type = ?", name).First(&good); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrGoodNotFound
		}

		return nil, result.Error
	}

	return &good, nil
}
