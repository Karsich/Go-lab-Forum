package handlers

import (
	"github.com/gin-gonic/gin"
	"main-service/db"
	"main-service/models"
	"net/http"
)

func GetTopics(c *gin.Context) {
	var topics []models.Topic
	if err := db.DB.Find(&topics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, topics)
}

func GetTopic(c *gin.Context) {
	topicID := c.Param("topic_id")
	var topic models.Topic
	if err := db.DB.Preload("Posts").First(&topic, topicID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}
	c.JSON(http.StatusOK, topic)
}

func GetPosts(c *gin.Context) {
	topicID := c.Param("topic_id")
	var posts []models.Post
	if err := db.DB.Where("topic_id = ?", topicID).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts)
}
