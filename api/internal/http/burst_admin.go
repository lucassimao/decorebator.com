package http

import (
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

// UnblockUser unblocks a specific user by ID (admin only)
func UnblockUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from URL parameter
		userIDStr := c.Param("userId")
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Get Redis client
		redisClient, err := common.GetRedisClient()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis connection failed"})
			return
		}

		// Create burst detector
		detector := service.NewBurstDetector(redisClient)

		// Check if user is actually blocked
		isBlocked := detector.IsUserBlocked(c.Request.Context(), userID)
		if !isBlocked {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User is not currently blocked",
				"user_id": userID,
			})
			return
		}

		// Unblock the user
		err = detector.UnblockUser(c.Request.Context(), userID)
		if err != nil {
			common.Logger.Error("Failed to unblock user", "error", err, "user_id", userID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unblock user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User unblocked successfully",
			"user_id": userID,
		})
	}
}

// UnblockAllUsers unblocks all currently blocked users (admin only)
func UnblockAllUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Redis client
		redisClient, err := common.GetRedisClient()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis connection failed"})
			return
		}

		// Create burst detector
		detector := service.NewBurstDetector(redisClient)

		// Unblock all users
		unlockedCount, err := detector.UnblockAllUsers(c.Request.Context())
		if err != nil {
			common.Logger.Error("Failed to unblock all users", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unblock all users"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":         "All users unblocked successfully",
			"unblocked_count": unlockedCount,
		})
	}
}

// GetBlockedUsers returns a list of all currently blocked users (admin only)
func GetBlockedUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Redis client
		redisClient, err := common.GetRedisClient()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis connection failed"})
			return
		}

		// Create burst detector
		detector := service.NewBurstDetector(redisClient)

		// Get blocked users
		blockedUsers, err := detector.GetBlockedUsers(c.Request.Context())
		if err != nil {
			common.Logger.Error("Failed to get blocked users", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve blocked users"})
			return
		}

		// Get detailed information for each blocked user
		var blockedUsersInfo []gin.H
		for _, userID := range blockedUsers {
			expiry := detector.GetBlockExpiry(c.Request.Context(), userID)
			userInfo := gin.H{
				"user_id": userID,
			}
			if expiry != nil {
				userInfo["blocked_until"] = expiry.Format("2006-01-02T15:04:05Z07:00")
			}
			blockedUsersInfo = append(blockedUsersInfo, userInfo)
		}

		c.JSON(http.StatusOK, gin.H{
			"blocked_users": blockedUsersInfo,
			"total_count":   len(blockedUsers),
		})
	}
}
