package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"main-service/models"
	"net/http"
)

func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userObj, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "User not found in context",
			})
			return
		}

		user, ok := userObj.(models.User)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user type in context",
			})
			return
		}

		if user.Role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":    fmt.Sprintf("%s access required", requiredRole),
				"required": requiredRole,
				"actual":   user.Role,
			})
			return
		}

		c.Next()
	}
}
