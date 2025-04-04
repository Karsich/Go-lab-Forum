package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"main-service/db"
	"main-service/handlers"
	"main-service/middleware"
)

func main() {
	err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)

	// Middleware для проверки JWT (заглушка)
	r.Use(func(c *gin.Context) {
		// Здесь должна быть проверка JWT токена
		c.Next()
	})

	// Public endpoints
	r.GET("/topics", handlers.GetTopics)
	r.GET("/topics/:topic_id", handlers.GetTopic)
	r.GET("/topics/:topic_id/posts", handlers.GetPosts)

	// Authenticated endpoints
	auth := r.Group("/")
	authMiddleware := middleware.AuthMiddleware(db.DB)
	auth.Use(authMiddleware)
	{
		auth.POST("/topics", handlers.CreateTopic)
		auth.POST("/topics/:topic_id/posts", handlers.CreatePost)
		auth.POST("/topics/:topic_id/posts/:post_id/replies", handlers.CreateReply)
		auth.POST("/topics/:topic_id/posts/:post_id/reactions", handlers.HandleReaction)
	}

	// Admin endpoints
	admin := r.Group("/")
	admin.Use(middleware.AuthMiddleware(db.DB))
	admin.Use(middleware.RoleMiddleware("admin"))
	{
		admin.DELETE("/topics/:topic_id", handlers.DeleteTopic)
		admin.DELETE("/topics/:topic_id/posts/:post_id", handlers.DeletePost)
		admin.PUT("/topics/:topic_id/status", handlers.UpdateTopicStatus)
	}

	log.Println("Forum service running on port 8082")
	if err := r.Run(":8082"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
