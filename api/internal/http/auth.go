package http

import (
	"net/http"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

// RefreshToken generates a new JWT token with updated user information
func RefreshToken(userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by auth middleware)
		userAny, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			return
		}
		user := userAny.(*model.User)

		// Fetch fresh user data from database to get updated subscription info
		users, err := userRepo.Find(repository.FindUserArgs{
			ID: &user.ID,
		})
		if err != nil || len(users) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data"})
			return
		}

		freshUser := users[0]

		// Generate new JWT with updated subscription info
		token, err := service.GenerateJWT(freshUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Return the new token
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":               freshUser.ID,
				"firstName":        freshUser.FirstName,
				"lastName":         freshUser.LastName,
				"email":            freshUser.Email,
				"subscriptionPlan": freshUser.SubscriptionPlan,
			},
		})
	}
}