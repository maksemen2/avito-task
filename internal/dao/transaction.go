package dao

import (
	"github.com/maksemen2/avito-shop/internal/models"
	"gorm.io/gorm"
)

type TransactionDAO struct {
	db *gorm.DB
}

func NewTransactionDAO(db *gorm.DB) *TransactionDAO {
	return &TransactionDAO{db: db}
}

// GetHistoryByUserID возвращает историю входящих и исходящих транзакций пользователя по его айди.
func (dao *TransactionDAO) GetHistoryByUserID(userID uint) ([]models.ReceivedCoins, []models.SentCoins, error) {
	var received []models.ReceivedCoins

	var sent []models.SentCoins

	// Полученные средства
	if err := dao.db.Table("transactions").
		Select("users.username as from_user, transactions.amount, transactions.created_at").
		Joins("JOIN users ON transactions.from_user_id = users.id").
		Where("transactions.to_user_id = ?", userID).
		Order("transactions.created_at DESC").
		Scan(&received).Error; err != nil {
		return nil, nil, err
	}

	// Отправленные средства
	if err := dao.db.Table("transactions").
		Select("users.username as to_user, transactions.amount, transactions.created_at").
		Joins("JOIN users ON transactions.to_user_id = users.id").
		Where("transactions.from_user_id = ?", userID).
		Order("transactions.created_at DESC").
		Scan(&sent).Error; err != nil {
		return nil, nil, err
	}

	if received == nil {
		received = make([]models.ReceivedCoins, 0)
	}

	if sent == nil {
		sent = make([]models.SentCoins, 0)
	}

	return received, sent, nil
}
