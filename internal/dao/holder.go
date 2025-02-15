package dao

import (
	"github.com/maksemen2/avito-shop/internal/database"
	"gorm.io/gorm"
)

type HolderDAO struct {
	DB          *gorm.DB
	User        *UserDAO
	Purchase    *PurchaseDAO
	Transaction *TransactionDAO
}

func NewHolderDAO(db *gorm.DB) *HolderDAO {
	return &HolderDAO{
		DB:          db,
		User:        NewUserDAO(db),
		Purchase:    NewPurchaseDAO(db),
		Transaction: NewTransactionDAO(db),
	}
}

func (dao *HolderDAO) TransferCoins(senderID, recieverID uint, amount int) error {

	tx := dao.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	//Вычитаем монеты у отправителя
	if err := tx.Model(&database.User{}).Where("id = ?", senderID).Update("coins", gorm.Expr("coins - ?", amount)).Error; err != nil {
		tx.Rollback()
		return err
	}

	//Добавляем получателю
	if err := tx.Model(&database.User{}).Where("id = ?", recieverID).Update("coins", gorm.Expr("coins + ?", amount)).Error; err != nil {
		tx.Rollback()
		return err
	}

	//Создаем запись о переводе монет
	if err := tx.Create(&database.Transaction{FromUserID: senderID, ToUserID: recieverID, Amount: amount}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (dao *HolderDAO) BuyItem(buyerID uint, goodName string) error {
	tx := dao.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	//Списываем монеты
	if err := tx.Model(&database.User{}).Where("id = ?", buyerID).Update("coins", gorm.Expr("coins - ?", 100)).Error; err != nil {
		tx.Rollback()
		return err
	}

	//Создаем запись о покупке
	if err := tx.Create(&database.Purchase{UserID: buyerID, ItemName: goodName}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
