package db

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"notification-service/models"
	"os"
)

var DB *gorm.DB

// Connect устанавливает подключение к базе данных
func Connect() {
	// Получаем строку подключения из переменной окружения или используем дефолтную
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=forum-postgres user=forum_admin password=strongpassword123 dbname=forum_db port=5432 sslmode=disable"
	}

	// Подключаемся к базе данных
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	// Присваиваем DB глобальной переменной
	DB = database
	fmt.Println("Подключение к БД успешно!")

	// Проверяем подключение
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Ошибка получения объекта sql.DB: %v", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("Ошибка при проверке подключения к базе данных: %v", err)
	}
	fmt.Println("Проверка подключения к базе данных прошла успешно!")
}

// GetUserEmailByID извлекает email пользователя по ID из базы данных
func GetUserEmailByID(userID int) (string, error) {
	// Проверка на случай, если база данных не подключена
	if DB == nil {
		return "", fmt.Errorf("database connection is nil")
	}

	var user models.User
	fmt.Println("Используем базу данных для запроса пользователя...")

	// Запрос к базе данных
	result := DB.First(&user, userID)
	if result.Error != nil {
		// Логирование ошибки
		log.Printf("Error fetching user by ID: %v", result.Error)
		return "", result.Error
	}

	// Возвращаем email пользователя
	return user.Email, nil
}

// SaveNotification сохраняет уведомление в базе данных
func SaveNotification(userID uint, notificationType string, message string) error {
	notification := models.Notification{
		UserID:  userID,
		Type:    notificationType,
		Message: message,
		Read:    false, // По умолчанию уведомление не прочитано.
	}

	// Сохраняем уведомление в базе данных
	if err := DB.Create(&notification).Error; err != nil {
		// Логируем ошибку, если не удалось сохранить уведомление
		log.Printf("Error saving notification for user %d: %v", userID, err)
		return err
	}

	// Логируем успешное сохранение
	log.Printf("Notification saved for user %d", userID)
	return nil
}

// Получение уведомлений по userID
func GetNotificationsByUserID(userID int) (string, error) {
	var notifications []models.Notification

	// Ищем уведомления для конкретного пользователя
	result := DB.Where("user_id = ?", userID).Find(&notifications)
	if result.Error != nil {
		return "", fmt.Errorf("error fetching notifications: %v", result.Error)
	}

	// Преобразуем в JSON
	notificationsJSON, err := json.Marshal(notifications)
	if err != nil {
		return "", fmt.Errorf("error marshaling notifications: %v", err)
	}

	return string(notificationsJSON), nil
}
