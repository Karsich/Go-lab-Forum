# Используем официальный образ golang
FROM golang:latest

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файл go.mod и go.sum для зависимостей
COPY go.mod ./
COPY go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем все файлы в контейнер
COPY . .

# Компилируем приложение
RUN go build -o notification-service .

# Открываем порт, на котором будет работать сервис
EXPOSE 8083

# Команда для запуска приложения
CMD ["./notification-service"]
