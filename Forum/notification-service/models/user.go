package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Username     string    `gorm:"unique"`
	Email        string    `gorm:"unique"`
	PasswordHash string    `gorm:"column:password_hash"`
	Role         string    `gorm:"default:user"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}
