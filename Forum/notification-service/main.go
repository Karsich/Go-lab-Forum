package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"notification-service/db"
	"notification-service/producer"
	"notification-service/worker"
	"strconv"
)

func main() {
	// Подключаемся к базе данных
	db.Connect()

	// Инициализация Redis
	worker.InitRedis()
	producer.InitRedis()

	// Запускаем воркер
	go worker.ProcessNotifications()

	// Регистрируем маршруты
	http.HandleFunc("/notifications", handleNotifications)
	http.HandleFunc("/send-notification", handleSendNotification)

	// Запускаем сервер
	log.Println("Notification service started on port 8083...")
	log.Fatal(http.ListenAndServe(":8083", nil))
}

// Обработчик POST запроса для отправки уведомлений
func handleSendNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Парсим JSON тело запроса
	var req struct {
		UserID  uint   `json:"user_id"`
		Type    string `json:"type"`
		Message string `json:"message"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Проверяем параметры
	if req.UserID == 0 || req.Type == "" || req.Message == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Отправляем уведомление
	err = producer.PublishNotification(uint64(req.UserID), req.Type, req.Message)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to publish notification: %v", err), http.StatusInternalServerError)
		return
	}

	// Ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Notification sent successfully"}`))
}

// Обработчик GET запроса для получения уведомлений по userID
func handleNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Получаем user_id из параметров запроса
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	// Получаем уведомления
	notifications, err := db.GetNotificationsByUserID(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get notifications: %v", err), http.StatusInternalServerError)
		return
	}

	// Отправляем JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(notifications))
}
