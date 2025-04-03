package worker

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"log"
	"net/smtp"
	"notification-service/db"
	"os"
	"strconv"
	"time"
)

var rdb *redis.Client

// Инициализация Redis-клиента
func InitRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "", // используйте пароль, если необходимо
		DB:       0,  // по умолчанию используем базу данных 0
	})

	// Проверка подключения к Redis
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_, err := rdb.Ping(ctx).Result()
		if err == nil {
			log.Println("Connected to Redis successfully!")
			break
		}
		log.Printf("Failed to connect to Redis, retrying... (%d/5)", i+1)
	}

	// Создание стрима и группы (если нет)
	err := rdb.XGroupCreateMkStream(ctx, "notifications", "notification-group", "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Fatalf("Failed to create stream or consumer group: %v", err)
	}
	log.Println("Stream and Consumer Group are ready")
}

// Функция для отправки email-уведомлений
func sendEmail(to string, message string) error {
	from := "dane.upton4@ethereal.email"
	pass := "YxVRsXNbTgkQahfkej"
	subject := "New Notification"
	body := fmt.Sprintf("Hello, you have a new notification: %s", message)
	log.Printf("Attempting to send email to: %s with message: %s", to, message)

	msg := []byte("From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body)

	err := smtp.SendMail(
		"smtp.ethereal.email:587",
		smtp.PlainAuth("", from, pass, "smtp.ethereal.email"),
		from, []string{to}, msg,
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent to: %s", to)
	return nil
}

// Обработка уведомлений
func processNotification(message redis.XMessage) {
	ctx := context.Background()

	// Извлекаем user_id из сообщения
	userIDStr, ok := message.Values["user_id"].(string)
	if !ok {
		log.Println("Failed to get user_id from message")
		return
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		log.Printf("Invalid user_id format: %v", err)
		return
	}

	// Извлекаем текст уведомления
	messageText, ok := message.Values["message"].(string)
	if !ok {
		log.Println("Failed to get message text from message")
		return
	}

	log.Printf("Processing notification for user %d", userID)

	// Получаем email пользователя из базы данных
	email, err := db.GetUserEmailByID(int(userID))
	if err != nil {
		log.Printf("Failed to get email for user %d: %v", userID, err)
		return
	}

	log.Printf("Sending notification to email: %s", email)

	// Отправляем уведомление на email
	err = sendEmail(email, messageText)
	if err != nil {
		log.Printf("Failed to send email to user %d: %v", userID, err)
	}

	// Сохраняем уведомление в базе данных
	err = db.SaveNotification(uint(userID), "new_message", messageText)
	if err != nil {
		log.Printf("Failed to save notification for user %d: %v", userID, err)
	}

	// Подтверждаем обработку сообщения
	_, err = rdb.XAck(ctx, "notifications", "notification-consumer", message.ID).Result()
	if err != nil {
		log.Printf("Failed to acknowledge message %s: %v", message.ID, err)
	}
}

// Основная функция для обработки уведомлений
func ProcessNotifications() {
	ctx := context.Background()

	for {
		// Чтение из Redis Stream
		streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    "notification-group",
			Consumer: "notification-consumer",
			Streams:  []string{"notifications", ">"},
			Count:    1,
			Block:    0, // Блокируемся до появления новых сообщений
		}).Result()

		if err != nil {
			log.Printf("Error reading from Redis Stream: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Обработка сообщений из потока
		for _, stream := range streams {
			for _, message := range stream.Messages {
				processNotification(message)
			}
		}
	}
}
