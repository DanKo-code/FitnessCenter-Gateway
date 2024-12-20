package middlewares

import (
	"Gateway/internal/common_middlewares/common_middlewares_errors"
	"Gateway/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

func IsAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("Role")
		if !exists {
			logger.ErrorLogger.Printf("Error getting Role: %v", common_middlewares_errors.RoleNotFoundInContext)
			c.JSON(http.StatusUnauthorized, gin.H{"error": common_middlewares_errors.RoleNotFoundInContext})
			return
		}

		if role != "admin" {
			logger.ErrorLogger.Printf("Error current user not admin: %v", common_middlewares_errors.CurrentUserNotAdmin)
			c.JSON(http.StatusForbidden, gin.H{"error": common_middlewares_errors.CurrentUserNotAdmin})
			return
		}

		c.Next()
	}
}
