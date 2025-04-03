package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"main-service/models"
	"os"
	"time"
)

var DB *gorm.DB

func InitDB() error {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=forum-postgres user=forum_admin password=strongpassword123 dbname=forum_db port=5432 sslmode=disable"
	}

	var err error
	for i := 0; i < 5; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Connection attempt %d failed: %v", i+1, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return err
	}

	// Автомиграция моделей
	err = DB.AutoMigrate(&models.User{}, &models.Topic{}, &models.Post{}, &models.Reaction{}, &models.Notification{}, &models.PrivateMessage{})
	if err != nil {
		return err
	}

	return nil
}
