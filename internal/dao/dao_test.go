package dao_test

import (
	"testing"

	"github.com/maksemen2/avito-shop/internal/dao"
	"github.com/maksemen2/avito-shop/internal/database"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func setupDao(t *testing.T) *dao.HolderDAO {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormLogger.Default.LogMode(gormLogger.Silent)})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	err = db.AutoMigrate(&database.User{}, &database.Purchase{}, &database.Transaction{}, &database.Good{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	goods := map[string]int{
		"t-shirt":    80,
		"cup":        20,
		"book":       50,
		"pen":        20,
		"powerbank":  200,
		"hoody":      300,
		"umbrella":   200,
		"socks":      10,
		"wallet":     50,
		"pink-hoody": 500,
	}

	for k, v := range goods {
		db.Create(&database.Good{Type: k, Price: v})
	}

	return dao.NewHolderDAO(db)
}

func TestTransferCoins(t *testing.T) {
	dao := setupDao(t)

	// Создаем пользователей: отправитель и получатель
	sender, err := dao.User.Create("sender", "12345678901234567890123456789012345678901234567890")
	assert.NoError(t, err, "sender creation error")
	receiver, err := dao.User.Create("receiver", "12345678901234567890123456789012345678901234567890")
	assert.NoError(t, err, "receiver creation error")

	// Проверяем начальный баланс
	assert.Equal(t, 1000, sender.Coins, "initial balance of user should be 1000 coins")
	assert.Equal(t, 1000, receiver.Coins, "initial balance of user should be 1000 coins")

	// Выполняем перевод 100 монет от отправителя к получателю
	err = dao.TransferCoins(sender.ID, receiver.ID, 100)
	assert.NoError(t, err, "error during coins transfer")

	// Обновляем данные пользователей
	sender, err = dao.User.GetByID(sender.ID)
	assert.NoError(t, err, "error getting sender")
	receiver, err = dao.User.GetByID(receiver.ID)
	assert.NoError(t, err, "error getting receiver")

	// Проверяем итоговый баланс
	assert.Equal(t, 900, sender.Coins, "sender should have 100 coins less")
	assert.Equal(t, 1100, receiver.Coins, "receiver should have 100 coins more")

	// Проверяем историю транзакций для отправителя
	recievedCoins, sentCoins, err := dao.Transaction.GetHistoryByUserID(sender.ID)
	assert.NoError(t, err, "error getting sender's transaction history")
	assert.Len(t, recievedCoins, 0, "sender should have no incoming transactions")
	assert.Len(t, sentCoins, 1, "sender should have one outgoing transaction")

	sent := sentCoins[0]
	assert.Equal(t, receiver.Username, sent.ToUser, "receiver's name in sent transaction is incorrect")
	assert.Equal(t, 100, sent.Amount, "transaction amount should be 100")

	// Проверяем историю транзакций для получателя
	recievedCoins, sentCoins, err = dao.Transaction.GetHistoryByUserID(receiver.ID)
	assert.NoError(t, err, "error getting receiver's transaction history")
	assert.Len(t, recievedCoins, 1, "receiver should have one incoming transaction")
	assert.Len(t, sentCoins, 0, "receiver should have no outgoing transactions")

	received := recievedCoins[0]
	assert.Equal(t, sender.Username, received.FromUser, "sender's name in received transaction is incorrect")
	assert.Equal(t, 100, received.Amount, "transaction amount should be 100")
}

func TestBuyItem(t *testing.T) {
	dao := setupDao(t)

	// Создаем пользователя
	user, err := dao.User.Create("user", "12345678901234567890123456789012345678901234567890")
	assert.NoError(t, err, "failed to create user")

	// Покупаем футболку за 80 монет
	tshirt, err := dao.Good.GetByName("t-shirt")
	assert.NoError(t, err, "failed to get tshirt")

	err = dao.BuyItem(user.ID, tshirt.ID, tshirt.Price)
	assert.NoError(t, err, "failed first t-shirt purchase")

	// Проверяем инвентарь
	inventory, err := dao.Purchase.GetInventoryByUserID(user.ID)
	assert.NoError(t, err, "failed to get inventory")
	assert.Len(t, inventory, 1, "expected one record in inventory")
	item := inventory[0]
	assert.Equal(t, "t-shirt", item.Type, "incorrect item type")
	assert.Equal(t, 1, item.Quantity, "expected quantity to be 1")

	// Проверяем баланс пользователя
	user, err = dao.User.GetByID(user.ID)
	assert.NoError(t, err, "failed to get user")
	assert.Equal(t, 1000-tshirt.Price, user.Coins, "user's balance should decrease by 80 coins")

	// Покупаем вторую футболку
	err = dao.BuyItem(user.ID, tshirt.ID, tshirt.Price)
	assert.NoError(t, err, "failed second t-shirt purchase")

	// Проверяем инвентарь после второй покупки
	inventory, err = dao.Purchase.GetInventoryByUserID(user.ID)
	assert.NoError(t, err, "failed to get inventory after second purchase")
	assert.Len(t, inventory, 1, "expected one record in inventory")
	item = inventory[0]
	assert.Equal(t, "t-shirt", item.Type, "incorrect item type")
	assert.Equal(t, 2, item.Quantity, "expected t-shirt quantity to be 2")

	// Проверяем баланс после второй покупки
	user, err = dao.User.GetByID(user.ID)
	assert.NoError(t, err, "failed to get user after second purchase")
	assert.Equal(t, 1000-tshirt.Price*2, user.Coins, "balance should further decrease by 80 coins")

	powerbank, err := dao.Good.GetByName("powerbank")
	assert.NoError(t, err, "failed to get powerbank")
	// Покупаем пауер банк за 200 монет
	err = dao.BuyItem(user.ID, powerbank.ID, powerbank.Price)
	assert.NoError(t, err, "failed to buy powerbank")

	// Проверяем инвентарь после покупки пауер банка
	inventory, err = dao.Purchase.GetInventoryByUserID(user.ID)
	assert.NoError(t, err, "failed to get inventory after powerbank purchase")
	assert.Len(t, inventory, 2, "expected inventory to have two entries")

	// Проверяем каждую запись инвентаря
	for _, i := range inventory {
		switch i.Type {
		case "powerbank":
			assert.Equal(t, 1, i.Quantity, "expected powerbank quantity to be 1")
		case "t-shirt":
			assert.Equal(t, 2, i.Quantity, "expected t-shirt quantity to be 2")
		}
	}

	// Получаем конечный баланс
	user, err = dao.User.GetByID(user.ID)
	assert.NoError(t, err, "failed to get final state of user")
	assert.Equal(t, 1000-tshirt.Price*2-powerbank.Price, user.Coins, "final user balance is incorrect")
}
