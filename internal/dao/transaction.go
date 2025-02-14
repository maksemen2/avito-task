package dao

import (
	"github.com/maksemen2/avito-task/internal/models"
	"gorm.io/gorm"
)

type TransactionDAO struct {
	db *gorm.DB
}

func NewTransactionDAO(db *gorm.DB) *TransactionDAO {
	return &TransactionDAO{db: db}
}

func (dao *TransactionDAO) GetHistoryByUserID(userID uint) ([]models.ReceivedCoins, []models.SentCoins, error) {
	var received []models.ReceivedCoins
	var sent []models.SentCoins

	if err := dao.db.Table("transactions t").
		Select("u.username as fromUser, t.amount").
		Joins("JOIN users u ON u.id = t.from_user_id").
		Where("t.to_user_id = ?", userID).
		Scan(&received).Error; err != nil {
		return nil, nil, err
	}

	if err := dao.db.Table("transactions t").
		Select("u.username as toUser, t.amount").
		Joins("JOIN users u ON u.id = t.to_user_id").
		Where("t.from_user_id = ?", userID).
		Scan(&sent).Error; err != nil {
		return nil, nil, err
	}

	if received == nil {
		received = []models.ReceivedCoins{}
	}
	if sent == nil {
		sent = []models.SentCoins{}
	}

	return received, sent, nil
}
