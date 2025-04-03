package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"notification-service/models"
	"os"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=forum-postgres user=forum_admin password=strongpassword123 dbname=forum_db port=5432 sslmode=disable"
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	DB = database
	fmt.Println("Подключение к БД успешно!")
}

func GetUserEmailByID(userID int) (string, error) {
	if DB == nil {
		return "", fmt.Errorf("database connection is nil")
	}

	var user models.User
	fmt.Println("Используем базу данных для запроса пользователя...")
	result := DB.First(&user, userID)
	if result.Error != nil {
		return "", result.Error
	}

	return user.Email, nil
}
