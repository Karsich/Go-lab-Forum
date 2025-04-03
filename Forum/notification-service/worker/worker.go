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

func InitRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})

	// Проверка подключения
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

func processNotification(message redis.XMessage) {
	ctx := context.Background()

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

	messageText, ok := message.Values["message"].(string)
	if !ok {
		log.Println("Failed to get message text from message")
		return
	}

	log.Printf("Processing notification for user %d", userID)

	// Получаем email из БД
	email, err := db.GetUserEmailByID(int(userID))
	if err != nil {
		log.Printf("Failed to get email for user %d: %v", userID, err)
		return
	}

	log.Printf("Testing email: Overriding user email to: %s", email)

	// Отправляем уведомление на email
	err = sendEmail(email, messageText)
	if err != nil {
		log.Printf("Failed to send email to user %d: %v", userID, err)
	}

	// Подтверждение обработки сообщения
	_, err = rdb.XAck(ctx, "notifications", "notification-consumer", message.ID).Result()
	if err != nil {
		log.Printf("Failed to acknowledge message %s: %v", message.ID, err)
	}
}

func ProcessNotifications() {
	ctx := context.Background()

	for {
		streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    "notification-group",
			Consumer: "notification-consumer",
			Streams:  []string{"notifications", ">"},
			Count:    1,
			Block:    0, // Ожидаем новые сообщения
		}).Result()

		if err != nil {
			log.Printf("Error reading from Redis Stream: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				processNotification(message)
			}
		}
	}
}
