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

	// Проверяем, есть ли ответы на это сообщение
	//var repliesCount int64
	//db.DB.Model(&models.Post{}).Where("parent_post_id = ?", postID).Count(&repliesCount)
	//if repliesCount > 0 {
	//	c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete post with replies"})
	//	return
	//}

	// Проверяем права (только автор может удалять)
	userID := c.GetUint("userID") // Предполагаем, что middleware установил userID
	var post models.Post
	if err := db.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	if post.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "No permission to delete this post"})
		return
	}

	if err := db.DB.Delete(&models.Post{}, postID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

func DeleteTopic(c *gin.Context) {
	id := c.Param("topic_id")

	// Каскадное удаление (тема + связанные посты + реакции)
	if err := db.DB.Select("Posts", "Posts.Reactions").Delete(&models.Topic{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Topic deleted successfully"})
}

func UpdateTopicStatus(c *gin.Context) {
	id := c.Param("topic_id")
	var topic models.Topic
	if err := db.DB.First(&topic, id).Error; err != nil {
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
	if err := db.DB.Save(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, topic)
}
