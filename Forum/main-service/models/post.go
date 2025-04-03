package models

import "time"

type Post struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Content      string    `gorm:"type:text;not null" json:"content"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	TopicID      uint      `gorm:"not null" json:"topic_id"`
	ParentPostID *uint     `gorm:"index;null" json:"parent_post_id,omitempty"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	User         *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ParentPost   *Post     `gorm:"foreignKey:ParentPostID" json:"-"`
	Replies      []Post    `gorm:"foreignKey:ParentPostID" json:"replies,omitempty"`
}
