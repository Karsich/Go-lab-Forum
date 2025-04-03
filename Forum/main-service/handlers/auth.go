package handlers

import (
	"github.com/gin-gonic/gin"
	"main-service/db"
	"main-service/models"
	"net/http"
	"strconv"
)

func CreateTopic(c *gin.Context) {
	var topic models.Topic
	if err := c.ShouldBindJSON(&topic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.DB.Create(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, topic)
}

func CreatePost(c *gin.Context) {
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
	var topic models.Topic
	if err := db.DB.First(&topic, topicID).Error; err != nil {
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

	post := models.Post{
		Content: input.Content,
		UserID:  userID.(uint), // Важно: приводим к типу uint
		TopicID: uint(topicID),
	}

	if err := db.DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func CreateReply(c *gin.Context) {
	// Получаем ID родительского сообщения
	parentPostIDStr := c.Param("post_id")
	parentPostID, err := strconv.ParseUint(parentPostIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Проверяем существование родительского сообщения
	var parentPost models.Post
	if err := db.DB.First(&parentPost, parentPostID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parent post not found"})
		return
	}

	// Проверяем, не закрыта ли тема
	var topic models.Topic
	if err := db.DB.First(&topic, parentPost.TopicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	if topic.Status == "closed" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot reply in closed topic"})
		return
	}

	// Получаем userID из middleware аутентификации
	userID := c.GetUint("userID")

	var reply models.Post
	if err := c.ShouldBindJSON(&reply); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Устанавливаем параметры ответа
	reply.UserID = userID
	reply.TopicID = parentPost.TopicID
	reply.ParentPostID = &parentPost.ID // Указываем родительское сообщение

	if err := db.DB.Create(&reply).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, reply)
}

func HandleReaction(c *gin.Context) {
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
	var post models.Post
	if err := db.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Ищем существующую реакцию
	var reaction models.Reaction
	dbResult := db.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&reaction)

	if input.Type == "" {
		// Удаление реакции
		if dbResult.Error == nil {
			if err := db.DB.Delete(&reaction).Error; err != nil {
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
			if err := db.DB.Save(&reaction).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			// Создаем новую реакцию
			reaction = models.Reaction{
				UserID: userID,
				PostID: uint(postID),
				Type:   input.Type,
			}
			if err := db.DB.Create(&reaction).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, reaction)
	}
}
