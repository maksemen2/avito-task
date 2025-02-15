package dao

import (
	"fmt"

	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/maksemen2/avito-shop/internal/models"
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

	transactionsTableName, usersTableName := database.Transaction{}.TableName(), database.User{}.TableName()

	if err := dao.db.Debug().Table(transactionsTableName).
		Select(fmt.Sprintf("%s.username as from_user, %s.amount", usersTableName, transactionsTableName)).
		Joins(fmt.Sprintf("JOIN %s ON %s.from_user_id = %s.id", usersTableName, transactionsTableName, usersTableName)).
		Where(fmt.Sprintf("%s.to_user_id = ?", transactionsTableName), userID).
		Scan(&received).Error; err != nil {
		return nil, nil, err
	}

	if err := dao.db.Debug().Table(transactionsTableName).
		Select(fmt.Sprintf("%s.username as to_user, %s.amount", usersTableName, transactionsTableName)).
		Joins(fmt.Sprintf("JOIN %s ON %s.to_user_id = %s.id", usersTableName, transactionsTableName, usersTableName)).
		Where(fmt.Sprintf("%s.from_user_id = ?", transactionsTableName), userID).
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
