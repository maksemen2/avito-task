package repository

import (
	"context"
	"fmt"

	"github.com/maksemen2/avito-shop/internal/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type HolderRepository interface {
	TransferCoins(ctx context.Context, senderID, receiverID uint, amount int) error
	BuyItem(ctx context.Context, buyerID, goodID uint, goodPrice int) error
	User() UserRepository
	Purchase() PurchaseRepository
	Transaction() TransactionRepository
	Good() GoodRepository
}

type GormHolderRepository struct {
	user        UserRepository
	purchase    PurchaseRepository
	transaction TransactionRepository
	good        GoodRepository
	logger      *zap.Logger
	BaseRepository
}

func NewHolderRepository(db *gorm.DB, logger *zap.Logger) HolderRepository {
	return &GormHolderRepository{
		user:        NewUserRepository(db, logger),
		purchase:    NewPurchaseRepository(db, logger),
		transaction: NewTransactionRepository(db, logger),
		good:        NewGoodRepository(db, logger),
		BaseRepository: BaseRepository{
			db:     db,
			Logger: logger,
		},
	}
}

// TransferCoins переводит монеты от одного пользователя к другому.
func (r *GormHolderRepository) TransferCoins(ctx context.Context, senderID, receiverID uint, amount int) error {
	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// Списываем баланс с дополнительной проверкой на его наличие
		res := tx.Model(&database.User{}).
			Where("id = ? AND coins >= ?", senderID, amount).
			UpdateColumn("coins", gorm.Expr("coins - ?", amount))

		if res.Error != nil {
			r.logger.Error("failed to transfer coins", zap.Uint("senderID", senderID), zap.Uint("recieverID", receiverID), zap.Error(res.Error))
			return WrapError(ErrTransferCoins.Error(), res.Error)
		}

		if res.RowsAffected == 0 {
			fmt.Println("no")
			// При валидации jwt токена мы можем верить что он создан именно сервером и не может быть подделан,
			// поэтому пользователь точно существует и проблема связана с недостатком средств
			return ErrInsufficientFunds
		}

		// Начисляем баланс получателю
		res = tx.Model(&database.User{}).
			Where("id = ?", receiverID).
			UpdateColumn("coins", gorm.Expr("coins + ?", amount))

		if res.Error != nil {
			r.logger.Error("failed to transfer coins", zap.Uint("senderID", senderID), zap.Uint("recieverID", receiverID), zap.Error(res.Error))
			return WrapError(ErrTransferCoins.Error(), res.Error)
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
			r.logger.Error("failed to create transaction", zap.Uint("senderID", senderID), zap.Uint("recieverID", receiverID), zap.Error(err))
			return WrapError(ErrTransferCoins.Error(), err)
		}

		return nil
	})
}

// BuyItem произовдит покупку товара пользователем.
func (r *GormHolderRepository) BuyItem(ctx context.Context, buyerID, goodID uint, goodPrice int) error {
	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// Списываем деньги и на редкий случай в котором количество монет на балансе изменилось в промежуток времени между проверкой в сервисе
		// и выполнением в этой транзакции выполняем проверку еще раз
		res := tx.Model(&database.User{}).
			Where("id = ? AND coins >= ?", buyerID, goodPrice).
			UpdateColumn("coins", gorm.Expr("coins - ?", goodPrice)) // Вряд ли цена товара изменится, да и функционала такого в проекте нет, поэтому можно себе позволить использовать уже полученную в сервисе цену

		if res.Error != nil {
			r.logger.Error("failed to buy item", zap.Uint("buyerID", buyerID), zap.Uint("goodID", goodID), zap.Error(res.Error))
			return WrapError(ErrBuyItem.Error(), res.Error)
		}

		if res.RowsAffected == 0 {
			// При валидации jwt токена мы можем верить что он создан именно сервером и не может быть подделан,
			// поэтому пользователь точно существует и проблема связана с недостатком средств
			return ErrInsufficientFunds
		}

		// Создаем запись о покупке
		if err := tx.Create(&database.Purchase{
			UserID: buyerID,
			GoodID: goodID,
		}).Error; err != nil {
			r.logger.Error("failed to create purchase", zap.Uint("buyerID", buyerID), zap.Uint("goodID", goodID), zap.Error(err))
			return WrapError(ErrBuyItem.Error(), err)
		}

		return nil
	})
}

func (r *GormHolderRepository) User() UserRepository {
	return r.user
}

func (r *GormHolderRepository) Purchase() PurchaseRepository {
	return r.purchase
}

func (r *GormHolderRepository) Transaction() TransactionRepository {
	return r.transaction
}

func (r *GormHolderRepository) Good() GoodRepository {
	return r.good
}
