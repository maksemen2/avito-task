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

// Нужно чтобы избежать хардкодинга имени таблицы в методе internal/dao/transaction.go/GetHistoryByUserID
func (Transaction) TableName() string {
	return "transactions"
}

type Purchase struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	User      *User `gorm:"foreignKey:UserID;references:ID"`
	ItemName  string
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
