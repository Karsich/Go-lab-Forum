package models

import (
	"time"
)

type PrivateMessage struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FromUserID uint      `gorm:"not null" json:"from_user_id"`
	ToUserID   uint      `gorm:"not null" json:"to_user_id"`
	Content    string    `gorm:"type:text" json:"content"`
	Read       bool      `gorm:"default:false"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}
