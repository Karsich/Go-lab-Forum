package main

import (
	"auth-service/db"
	"auth-service/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	db.Connect()

	r := gin.Default()

	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	r.Run(":8081")
}
