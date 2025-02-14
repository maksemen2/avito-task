package database

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"unique"`
	PasswordHash string
	Coins        int       `gorm:"default:1000"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

type Transaction struct {
	gorm.Model
	ID         uint `gorm:"primaryKey"`
	FromUserID uint
	FromUser   *User `gorm:"foreignKey:FromUserID;references:ID"`
	ToUserID   uint
	ToUser     *User `gorm:"foreignKey:ToUserID;references:ID"`
	Amount     int
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

type Purchase struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	User      *User `gorm:"foreignKey:UserID;references:ID"`
	ItemName  string
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
