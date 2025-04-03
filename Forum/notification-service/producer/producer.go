package producer

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"log"
	"os"
)

var rdb *redis.Client

func InitRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})

	// Проверка подключения
	for i := 0; i < 5; i++ { // Пытаемся 5 раз подключиться
		_, err := rdb.Ping(context.Background()).Result()
		if err == nil {
			log.Println("Connected to Redis successfully!")
			return
		}
		log.Printf("Failed to connect to Redis, retrying... (%d/5)", i+1)
	}
	log.Fatal("Could not connect to Redis after 5 attempts")
}

func PublishNotification(userID uint64, notificationType string, message string) error {
	ctx := context.Background()

	// Формируем данные для уведомления
	data := map[string]interface{}{
		"user_id": fmt.Sprintf("%d", userID),
		"message": message,
		"type":    notificationType,
	}

	// Публикуем в Redis Stream
	_, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "notifications",
		Values: data,
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to publish notification: %w", err)
	}
	return nil
}
