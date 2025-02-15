package dao_test

import (
	"fmt"
	"testing"

	"github.com/maksemen2/avito-shop/internal/dao"
	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDao(t *testing.T) *dao.HolderDAO {
	// Используем SQLite бд в оперативной памяти для тестов
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err, "некорректное открытие тестовой БД")

	err = db.AutoMigrate(&database.User{}, &database.Purchase{}, &database.Transaction{})
	assert.NoError(t, err, "ошибка миграции тестовой БД")
	return dao.NewHolderDAO(db)
}

func TestTransferCoins(t *testing.T) {
	dao := setupDao(t)

	sender, err := dao.User.Create("sender", "12345678901234567890123456789012345678901234567890")
	assert.NoError(t, err, "пользователь должен быть создан успешно")
	reciever, err := dao.User.Create("reciever", "12345678901234567890123456789012345678901234567890")
	assert.NoError(t, err, "пользователь должен быть создан успешно")

	assert.Equal(t, 1000, sender.Coins, "значение по умолчанию должно автоматически подставиться")
	assert.Equal(t, 1000, reciever.Coins, "значение по умолчанию должно автоматически подставиться")

	assert.NoError(t, dao.TransferCoins(sender.ID, reciever.ID, 100), "перевод монет должен пройти успешно")

	sender, err = dao.User.GetByID(sender.ID)
	assert.NoError(t, err, "пользователь должен быть получен успешно")

	reciever, err = dao.User.GetByID(reciever.ID)
	assert.NoError(t, err, "пользователь должен быть получен успешно")

	assert.Equal(t, 900, sender.Coins, "у отправителя должно стать на 100 монет меньше")
	assert.Equal(t, 1100, reciever.Coins, "у получателя должно стать на 100 монет больше")

	// Проверяем, что запись о переводе создана и информация о входящих и исходящих переводах будет корректной

	recievedCoins, sentCoins, err := dao.Transaction.GetHistoryByUserID(sender.ID)
	fmt.Println(recievedCoins, sentCoins)
	assert.NoError(t, err, "история переводов должна быть получена успешно")
	assert.Len(t, recievedCoins, 0, "у отправителя не должно быть входящих переводов")
	assert.Len(t, sentCoins, 1, "у отправителя должен быть один исходящий перевод")
	sent := sentCoins[0]
	assert.Equal(t, reciever.Username, sent.ToUser, "имя получателя должно совпадать с записью")
	assert.Equal(t, 100, sent.Amount, "сумма перевода должна быть равна 100")

	recievedCoins, sentCoins, err = dao.Transaction.GetHistoryByUserID(reciever.ID)
	assert.NoError(t, err, "история переводов должна быть получена успешно")
	assert.Len(t, recievedCoins, 1, "у получателя должен быть один входящий перевод")
	assert.Len(t, sentCoins, 0, "у получателя не должно быть исходящих переводов")
	recieved := recievedCoins[0]
	assert.Equal(t, sender.Username, recieved.FromUser, "имя отправителя должно совпадать с записью")
	assert.Equal(t, 100, recieved.Amount, "сумма перевода должна быть равна 100")
}

func TestBuyItem(t *testing.T) {
	dao := setupDao(t)

	user, err := dao.User.Create("user", "12345678901234567890123456789012345678901234567890")
	assert.NoError(t, err, "пользователь должен быть создан успешно")

	// Цена футболки - 80 монет
	err = dao.BuyItem(user.ID, "t-shirt", 80)
	assert.NoError(t, err, "покупка должна пройти успешно")

	inventory, err := dao.Purchase.GetInventoryByUserID(user.ID)
	assert.NoError(t, err, "инвентарь должен быть получен успешно")

	assert.Len(t, inventory, 1, "в инвентаре должна быть одна запись")
	item := inventory[0]

	assert.Equal(t, "t-shirt", item.Type, "имя товара должно совпадать с записью")
	assert.Equal(t, 1, item.Quantity, "количество товара должно быть равно 1")
	user, err = dao.User.GetByID(user.ID)

	assert.NoError(t, err, "пользователь должен быть получен успешно")
	assert.Equal(t, 920, user.Coins, "у пользователя должно стать на 80 монет меньше")

	// Купим еще одну футболку
	err = dao.BuyItem(user.ID, "t-shirt", 80)
	assert.NoError(t, err, "покупка должна пройти успешно")

	inventory, err = dao.Purchase.GetInventoryByUserID(user.ID)
	assert.NoError(t, err, "инвентарь должен быть получен успешно")

	assert.Len(t, inventory, 1, "в инвентаре должна остаться одна запись")
	item = inventory[0]

	assert.Equal(t, "t-shirt", item.Type, "имя товара должно совпадать с записью")
	assert.Equal(t, 2, item.Quantity, "количество товара должно быть равно 2")

	user, err = dao.User.GetByID(user.ID)
	assert.NoError(t, err, "пользователь должен быть получен успешно")
	assert.Equal(t, 840, user.Coins, "у пользователя должно стать на 80 монет меньше")

	// Купим еще пауер-банк ценой 200 монет
	err = dao.BuyItem(user.ID, "powerbank", 200)
	assert.NoError(t, err, "покупка должна пройти успешно")
	inventory, err = dao.Purchase.GetInventoryByUserID(user.ID)
	assert.NoError(t, err, "инвентарь должен быть получен успешно")
	assert.Len(t, inventory, 2, "в инвентаре должно быть две записи")
	for _, item := range inventory {
		if item.Type == "powerbank" {
			assert.Equal(t, 1, item.Quantity, "количество пауер-банков должно быть равно 1")
		} else if item.Type == "t-shirt" {
			assert.Equal(t, 2, item.Quantity, "количество футболок должно быть равно 2")
		}
	}

	user, err = dao.User.GetByID(user.ID)

	assert.NoError(t, err, "пользователь должен быть получен успешно")

	assert.Equal(t, 640, user.Coins, "у пользователя должно стать еще на 200 монет меньше")
}
