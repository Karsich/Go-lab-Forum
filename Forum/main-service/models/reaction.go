package models

import "time"

type Reaction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Type      string    `gorm:"size:50;not null" json:"type"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	PostID    uint      `gorm:"not null" json:"post_id"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Post      *Post     `gorm:"foreignKey:PostID" json:"-"`
}
