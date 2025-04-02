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
	gorm.Model
	Username     string `gorm:"unique;not null"`
	Email        string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"default:user"`
}

type Topic struct {
	gorm.Model
	Title       string
	Description string
	UserID      uint
	Status      string `gorm:"default:open"`
	Posts       []Post
}

type Post struct {
	gorm.Model
	Content    string
	UserID     uint
	TopicID    uint
	ParentPost *uint `gorm:"default:null"`
	Reactions  []Reaction
}

type Reaction struct {
	gorm.Model
	Type   string
	UserID uint
	PostID uint
}

var db *gorm.DB

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
		auth.POST("/topics/:topic_id/posts/:post_id/reactions", addReaction)
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
	// В реальности проверяем JWT токен
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
	topicIDStr := c.Param("topic_id")
	topicID, err := strconv.ParseUint(topicIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic ID"})
		return
	}

	var post Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post.TopicID = uint(topicID)
	if err := db.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func addReaction(c *gin.Context) {
	topicID := c.Param("topic_id")
	postID := c.Param("post_id")

	// Проверка существования темы и поста
	var topic Topic
	if err := db.First(&topic, topicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	var post Post
	if err := db.Where("id = ? AND topic_id = ?", postID, topicID).First(&post).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found in this topic"})
		return
	}

	var reaction Reaction
	if err := c.ShouldBindJSON(&reaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reaction.PostID = post.ID
	if err := db.Create(&reaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, reaction)
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
