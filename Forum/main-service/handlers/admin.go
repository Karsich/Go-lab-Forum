package handlers

import (
	"github.com/gin-gonic/gin"
	"main-service/db"
	"main-service/models"
	"net/http"
	"strconv"
)

func DeletePost(c *gin.Context) {
	postIDStr := c.Param("post_id")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	currentUser := c.MustGet("user").(models.User)

	var post models.Post
	if err := db.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Проверка прав: админ или автор поста
	if currentUser.Role != "admin" && post.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{
			"error":    "No permission to delete this post",
			"required": "admin or post author",
		})
		return
	}

	// Удаляем только этот пост (ответы остаются)
	if err := db.DB.Delete(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

func DeleteTopic(c *gin.Context) {
	topicID := c.Param("topic_id")

	currentUser := c.MustGet("user").(models.User)

	var topic models.Topic
	if err := db.DB.First(&topic, topicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	// Проверка прав: только админ может удалять темы
	if currentUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error":    "Admin access required",
			"required": "admin",
		})
		return
	}

	// Каскадное удаление темы и всех связанных постов
	if err := db.DB.Select("Posts").Delete(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Topic and all posts deleted successfully"})
}

func UpdateTopicStatus(c *gin.Context) {
	topicID := c.Param("topic_id")

	currentUser := c.MustGet("user").(models.User)

	var topic models.Topic
	if err := db.DB.First(&topic, topicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	// Проверка прав: админ или автор темы
	if currentUser.Role != "admin" && topic.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{
			"error":    "No permission to update this topic",
			"required": "admin or topic author",
		})
		return
	}

	var input struct {
		Status string `json:"status" binding:"required,oneof=open closed"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	topic.Status = input.Status
	if err := db.DB.Save(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Topic status updated",
		"status":  topic.Status,
	})
}
