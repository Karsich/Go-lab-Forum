package handlers

import (
	"github.com/gin-gonic/gin"
	"main-service/db"
	"main-service/models"
	"net/http"
	"strconv"
)

func CreateTopic(c *gin.Context) {
	// Получаем пользователя из контекста (установлено в middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var input struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	topic := models.Topic{
		Title:       input.Title,
		Description: input.Description,
		UserID:      userID.(uint), // Назначаем автором текущего пользователя
		Status:      "open",        // Статус по умолчанию
	}

	if err := db.DB.Create(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, topic)
}

func CreatePost(c *gin.Context) {
	// Проверка аутентификации
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	topicIDStr := c.Param("topic_id")
	topicID, err := strconv.ParseUint(topicIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid topic ID"})
		return
	}

	// Проверка существования и доступности темы
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
		UserID:  userID.(uint),
		TopicID: uint(topicID),
	}

	if err := db.DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func CreateReply(c *gin.Context) {
	// Проверка аутентификации
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	parentPostIDStr := c.Param("post_id")
	parentPostID, err := strconv.ParseUint(parentPostIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Проверка родительского сообщения
	var parentPost models.Post
	if err := db.DB.First(&parentPost, parentPostID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parent post not found"})
		return
	}

	// Проверка темы
	var topic models.Topic
	if err := db.DB.First(&topic, parentPost.TopicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	if topic.Status == "closed" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot reply in closed topic"})
		return
	}

	var input struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reply := models.Post{
		Content:      input.Content,
		UserID:       userID.(uint),
		TopicID:      parentPost.TopicID,
		ParentPostID: &parentPost.ID,
	}

	if err := db.DB.Create(&reply).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, reply)
}

func HandleReaction(c *gin.Context) {
	// Проверка аутентификации
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	postIDStr := c.Param("post_id")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var input struct {
		Type string `json:"type" binding:"omitempty,oneof=like dislike"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка существования поста
	var post models.Post
	if err := db.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Поиск существующей реакции
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
		// Добавление/изменение реакции
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
			reaction = models.Reaction{
				UserID: userID.(uint),
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
