package dao

import (
	"fmt"

	"github.com/maksemen2/avito-shop/internal/database"
	"gorm.io/gorm"
)

type HolderDAO struct {
	DB          *gorm.DB
	User        *UserDAO
	Purchase    *PurchaseDAO
	Transaction *TransactionDAO
	Good        *GoodDAO
}

func NewHolderDAO(db *gorm.DB) *HolderDAO {
	return &HolderDAO{
		DB:          db,
		User:        NewUserDAO(db),
		Purchase:    NewPurchaseDAO(db),
		Transaction: NewTransactionDAO(db),
		Good:        NewGoodDAO(db),
	}
}

// TransferCoins переводит монеты от одного пользователя к другому.
func (dao *HolderDAO) TransferCoins(senderID, receiverID uint, amount int) error {
	return dao.DB.Transaction(func(tx *gorm.DB) error {
		// Списываем баланс с дополнительной проверкой на его наличие
		res := tx.Model(&database.User{}).
			Where("id = ? AND coins >= ?", senderID, amount).
			UpdateColumn("coins", gorm.Expr("coins - ?", amount))

		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected == 0 {
			return ErrInsufficientFunds // Проверка на существование пользователя проводится в хендлере, соответственно ошибка точно связана с отсутствием средств
		}

		// Начисляем баланс получателю
		res = tx.Model(&database.User{}).
			Where("id = ?", receiverID).
			UpdateColumn("coins", gorm.Expr("coins + ?", amount))

		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected == 0 {
			return ErrUserNotFound
		}

		// Создаем запись о переводе
		if err := tx.Create(&database.Transaction{
			FromUserID: senderID,
			ToUserID:   receiverID,
			Amount:     amount,
		}).Error; err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		return nil
	})
}

// BuyItem произовдит покупку товара пользователем.
func (dao *HolderDAO) BuyItem(buyerID uint, goodID uint, goodPrice int) error {
	return dao.DB.Transaction(func(tx *gorm.DB) error {
		// Списываем деньги и на редкий случай в котором количество монет на балансе изменилось в промежуток времени между проверкой в хендлере
		// и выполнением в этой транзакции выполняем проверку еще раз
		res := tx.Model(&database.User{}).
			Where("id = ? AND coins >= ?", buyerID, goodPrice).
			UpdateColumn("coins", gorm.Expr("coins - ?", goodPrice)) // Вряд ли цена товара изменится, да и функционала такого в проекте нет, поэтому можно себе позволить использовать уже полученную в хендлере цену

		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected == 0 {
			// Проверка на существование пользователя проводится в хендлере, соответственно, если ни одна запись не была
			// изменена - мы можем быть уверены, что дело в недостатке средств
			return ErrInsufficientFunds
		}

		// Создаем запись о покупке
		if err := tx.Create(&database.Purchase{
			UserID: buyerID,
			GoodID: goodID,
		}).Error; err != nil {
			return fmt.Errorf("failed to create purchase: %w", err)
		}

		return nil
	})
}
