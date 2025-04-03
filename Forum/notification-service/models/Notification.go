package models

type Notification struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	UserID  uint   `gorm:"not null" json:"user_id"`
	Type    string `gorm:"size:50" json:"type"`
	Message string `gorm:"type:text" json:"message"`
	Read    bool   `gorm:"default:false"`
}
