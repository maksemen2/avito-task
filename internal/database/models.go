package database

import (
	"time"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"unique"`
	PasswordHash string
	Coins        int       `gorm:"default:1000"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

type Transaction struct {
	ID         uint `gorm:"primaryKey"`
	FromUserID uint
	FromUser   *User `gorm:"foreignKey:FromUserID;references:ID"`
	ToUserID   uint
	ToUser     *User `gorm:"foreignKey:ToUserID;references:ID"`
	Amount     int
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

// Необходимо для избежания хардкодинга имен таблиц в internal/dao/transaction.go/GetHistoryByUserID, маппер при создании таблиц будет использовать имена таблиц из этих методов
func (Transaction) TableName() string {
	return "transactions"
}

func (User) TableName() string {
	return "users"
}

type Purchase struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	User      *User `gorm:"foreignKey:UserID;references:ID"`
	ItemName  string
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
