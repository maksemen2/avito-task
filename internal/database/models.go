package database

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Username     string    `gorm:"uniqueIndex;size:255"`
	PasswordHash string    `gorm:"type:char(60)"`
	Coins        int       `gorm:"default:1000;check:coins >= 0"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

type Transaction struct {
	ID         uint      `gorm:"primaryKey;index"`
	FromUserID uint      `gorm:"index"`
	FromUser   *User     `gorm:"foreignKey:FromUserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ToUserID   uint      `gorm:"index"`
	ToUser     *User     `gorm:"foreignKey:ToUserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Amount     int       `gorm:"check:amount > 0"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

type Good struct {
	ID    uint   `gorm:"primaryKey"`
	Type  string `gorm:"uniqueIndex;size:255"`
	Price int    `gorm:"check:price > 0"`
}

type Purchase struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	User      *User     `gorm:"foreignKey:UserID"`
	GoodID    uint      `gorm:"index"`
	Good      *Good     `gorm:"foreignKey:GoodID"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
