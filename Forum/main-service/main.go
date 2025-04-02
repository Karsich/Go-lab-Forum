package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Модели данных
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:100;unique;not null" json:"username"`
	Email        string    `gorm:"size:100;unique;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Role         string    `gorm:"size:50;default:user" json:"role"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type Topic struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:255;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	Status      string    `gorm:"size:50;default:open" json:"status"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Posts       []Post    `gorm:"foreignKey:TopicID" json:"posts,omitempty"`
}

type Post struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Content      string    `gorm:"type:text;not null" json:"content"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	TopicID      uint      `gorm:"not null" json:"topic_id"`
	ParentPostID *uint     `gorm:"index;null" json:"parent_post_id,omitempty"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	User         *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ParentPost   *Post     `gorm:"foreignKey:ParentPostID" json:"-"`
	Replies      []Post    `gorm:"foreignKey:ParentPostID" json:"replies,omitempty"`
}

type Reaction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Type      string    `gorm:"size:50;not null" json:"type"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	PostID    uint      `gorm:"not null" json:"post_id"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Post      *Post     `gorm:"foreignKey:PostID" json:"-"`
}

var db *gorm.DB

func createTestUser() {
	testUser := User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password_123",
		Role:         "user",
	}
	db.FirstOrCreate(&testUser, User{Username: "testuser"})
}

func initDB() error {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=forum-postgres user=forum_admin password=strongpassword123 dbname=forum_db port=5432 sslmode=disable"
	}

	var err error
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
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
	err = db.AutoMigrate(&User{}, &Topic{}, &Post{}, &Reaction{})
	if err != nil {
		return err
	}

	createTestUser()
	return nil
}

func main() {
	// Инициализация БД
	err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Инициализация роутера
	r := gin.Default()
	r.SetTrustedProxies(nil) // Важно для безопасности

	// Middleware для проверки JWT (заглушка)
	r.Use(func(c *gin.Context) {
		// Здесь должна быть проверка JWT токена
		c.Next()
	})

	// Public endpoints
	r.GET("/topics", getTopics)
	r.GET("/topics/:topic_id", getTopic)
	r.GET("/topics/:topic_id/posts", getPosts)

	// Authenticated endpoints
	auth := r.Group("/")
	auth.Use(authMiddleware)
	{
		auth.POST("/topics", createTopic)
		auth.POST("/topics/:topic_id/posts", createPost)
		auth.POST("/topics/:topic_id/posts/:post_id/replies", createReply)
		auth.DELETE("/topics/:topic_id/posts/:post_id", deletePost)
		auth.POST("/topics/:topic_id/posts/:post_id/reactions", handleReaction)
	}

	// Admin endpoints
	admin := r.Group("/")
	admin.Use(adminMiddleware)
	{
		admin.DELETE("/topics/:topic_id", deleteTopic)
		admin.PUT("/topics/:topic_id/status", updateTopicStatus)
	}

	log.Println("Forum service running on port 8082")
	if err := r.Run(":8082"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Middleware (заглушки)
func authMiddleware(c *gin.Context) {
	// Временная заглушка - используем ID тестового пользователя
	c.Set("userID", uint(1))
	c.Next()
}

func adminMiddleware(c *gin.Context) {
	// В реальности проверяем роль пользователя
	c.Next()
}

// Обработчики запросов
func getTopics(c *gin.Context) {
	var topics []Topic
	if err := db.Find(&topics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, topics)
}

func getTopic(c *gin.Context) {
	topicID := c.Param("topic_id")
	var topic Topic
	if err := db.Preload("Posts").First(&topic, topicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}
	c.JSON(http.StatusOK, topic)
}

func getPosts(c *gin.Context) {
	topicID := c.Param("topic_id")
	var posts []Post
	if err := db.Where("topic_id = ?", topicID).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts)
}

func createTopic(c *gin.Context) {
	var topic Topic
	if err := c.ShouldBindJSON(&topic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Create(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, topic)
}

func createPost(c *gin.Context) {
	// Получаем userID из middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	topicIDStr := c.Param("topic_id")
	topicID, err := strconv.ParseUint(topicIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic ID"})
		return
	}

	// Проверяем существование темы
	var topic Topic
	if err := db.First(&topic, topicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	if topic.Status == "closed" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot post in closed topic"})
		return
	}

	var input struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post := Post{
		Content: input.Content,
		UserID:  userID.(uint), // Важно: приводим к типу uint
		TopicID: uint(topicID),
	}

	if err := db.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func createReply(c *gin.Context) {
	// Получаем ID родительского сообщения
	parentPostIDStr := c.Param("post_id")
	parentPostID, err := strconv.ParseUint(parentPostIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Проверяем существование родительского сообщения
	var parentPost Post
	if err := db.First(&parentPost, parentPostID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parent post not found"})
		return
	}

	// Проверяем, не закрыта ли тема
	var topic Topic
	if err := db.First(&topic, parentPost.TopicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	if topic.Status == "closed" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot reply in closed topic"})
		return
	}

	// Получаем userID из middleware аутентификации
	userID := c.GetUint("userID")

	var reply Post
	if err := c.ShouldBindJSON(&reply); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Устанавливаем параметры ответа
	reply.UserID = userID
	reply.TopicID = parentPost.TopicID
	reply.ParentPostID = &parentPost.ID // Указываем родительское сообщение

	if err := db.Create(&reply).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, reply)
}

func deletePost(c *gin.Context) {
	postIDStr := c.Param("post_id")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Проверяем, есть ли ответы на это сообщение
	var repliesCount int64
	db.Model(&Post{}).Where("parent_post_id = ?", postID).Count(&repliesCount)
	if repliesCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete post with replies"})
		return
	}

	// Проверяем права (только автор может удалять)
	userID := c.GetUint("userID") // Предполагаем, что middleware установил userID
	var post Post
	if err := db.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	if post.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "No permission to delete this post"})
		return
	}

	if err := db.Delete(&Post{}, postID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

func handleReaction(c *gin.Context) {
	// Получаем ID поста
	postIDStr := c.Param("post_id")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Получаем ID пользователя из middleware
	userID := c.GetUint("userID")

	// Парсим входные данные
	var input struct {
		Type string `json:"type"` // Может быть "like", "dislike" или пустая строка для удаления
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем существование поста
	var post Post
	if err := db.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Ищем существующую реакцию
	var reaction Reaction
	dbResult := db.Where("user_id = ? AND post_id = ?", userID, postID).First(&reaction)

	if input.Type == "" {
		// Удаление реакции
		if dbResult.Error == nil {
			if err := db.Delete(&reaction).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Reaction removed"})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"message": "No reaction to remove"})
		}
	} else {
		if dbResult.Error == nil {
			if reaction.Type == input.Type {
				c.JSON(http.StatusOK, gin.H{"message": "Reaction already exists", "reaction": reaction})
				return
			}
			reaction.Type = input.Type
			if err := db.Save(&reaction).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			// Создаем новую реакцию
			reaction = Reaction{
				UserID: userID,
				PostID: uint(postID),
				Type:   input.Type,
			}
			if err := db.Create(&reaction).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, reaction)
	}
}

func deleteTopic(c *gin.Context) {
	id := c.Param("topic_id")

	// Каскадное удаление (тема + связанные посты + реакции)
	if err := db.Select("Posts", "Posts.Reactions").Delete(&Topic{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Topic deleted successfully"})
}

func updateTopicStatus(c *gin.Context) {
	id := c.Param("topic_id")
	var topic Topic
	if err := db.First(&topic, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	topic.Status = input.Status
	if err := db.Save(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, topic)
}
