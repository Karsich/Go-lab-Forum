package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"notification-service/db"
	"notification-service/producer"
	"notification-service/worker"
)

func main() {
	// Настроим обработку сигналов для корректного завершения работы
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Подключаемся к базе данных
	db.Connect()

	// Инициализация Redis
	worker.InitRedis()
	producer.InitRedis()

	// Запускаем воркер для обработки уведомлений
	go worker.ProcessNotifications()

	// Пример: отправка уведомлений через продюсер
	go func() {
		// Допустим, мы отправляем уведомление пользователю с ID 1
		err := producer.PublishNotification(1, "new_message", "You have a new message on the forum.")
		if err != nil {
			log.Printf("Error publishing notification: %v", err)
		} else {
			log.Println("Notification published successfully.")
		}
	}()

	// Ожидаем завершения работы
	<-sigs
	fmt.Println("Notification service stopping...")
}
