package http

import (
	"io"
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

// CreateCheckoutSessionRequest represents the request to create a checkout session
type CreateCheckoutSessionRequest struct {
	Plan    string `json:"plan" binding:"required,oneof=monthly annual"`
	ExpoUri string `json:"expoUri" binding:"required"`
}

// CreateCheckoutSession creates a Stripe checkout session
func CreateCheckoutSession(subService *service.SubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by auth middleware)
		userAny, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			return
		}
		user, ok := userAny.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
			return
		}

		// Parse request
		var req CreateCheckoutSessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Convert plan string to model.SubscriptionPlan
		var plan model.SubscriptionPlan
		switch req.Plan {
		case "monthly":
			plan = model.PlanMonthly
		case "annual":
			plan = model.PlanAnnual
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan"})
			return
		}

		// Create checkout session
		session, err := subService.CreateCheckoutSession(c.Request.Context(), user.ID, user.Email, plan, req.ExpoUri)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"checkoutUrl": session.URL,
			"sessionId":   session.ID,
		})
	}
}

// HandleStripeWebhook handles Stripe webhook events
func HandleStripeWebhook(subService *service.SubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the request body
		payload, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		// Get Stripe signature from header
		signature := c.GetHeader("Stripe-Signature")
		if signature == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Stripe signature"})
			return
		}

		// Handle the webhook
		if err := subService.HandleWebhook(c.Request.Context(), payload, signature); err != nil {
			// Log the error but return 200 to Stripe to avoid retries
			common.Logger.Error("Failed to handle Stripe webhook", "error", err)
			// Return 200 to acknowledge receipt and prevent retries
			c.JSON(http.StatusOK, gin.H{"received": true})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

// GetSubscriptionStatus returns the user's current subscription status
func GetSubscriptionStatus(subRepo *repository.SubscriptionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userAny, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			return
		}
		user, ok := userAny.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
			return
		}

		// Get active subscription
		subscription, err := subRepo.GetActiveSubscriptionForUser(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription"})
			return
		}

		// Return subscription info
		response := gin.H{
			"plan":   user.SubscriptionPlan,
			"status": user.SubscriptionStatus,
		}

		if subscription != nil {
			response["currentPeriodEnd"] = subscription.CurrentPeriodEnd
			response["cancelAtPeriodEnd"] = subscription.CancelAtPeriodEnd
			response["trialEnd"] = subscription.TrialEnd
			response["plan"] = subscription.Plan
			response["status"] = subscription.Status
		}

		c.JSON(http.StatusOK, response)
	}
}

// CancelSubscription cancels the user's subscription
func CancelSubscription(subService *service.SubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userAny, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			return
		}
		user, ok := userAny.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
			return
		}

		// Cancel subscription
		if err := subService.CancelSubscription(c.Request.Context(), user.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Subscription will be cancelled at the end of the billing period"})
	}
}

// GetSubscriptionHistory returns the user's subscription history
func GetSubscriptionHistory(subRepo *repository.SubscriptionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userAny, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			return
		}
		user, ok := userAny.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
			return
		}

		// Get subscription history
		subscriptions, err := subRepo.GetUserSubscriptionHistory(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription history"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"subscriptions": subscriptions})
	}
}

// CheckSubscriptionLimits middleware checks if user has permission based on subscription
func CheckSubscriptionLimits(subService *service.SubscriptionService, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userAny, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			c.Abort()
			return
		}
		user, ok := userAny.(*model.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
			return
		}

		// For word operations, we might need wordlist ID
		if action == "add_word" {
			// Get wordlist ID from path or query
			wordlistIDStr := c.Param("wordlistId")
			if wordlistIDStr == "" {
				wordlistIDStr = c.Query("wordlistId")
			}
			if wordlistIDStr == "" {
				// Try to get from request body
				var body map[string]interface{}
				if err := c.ShouldBindJSON(&body); err == nil {
					if wlID, ok := body["wordlistId"].(float64); ok {
						wordlistIDStr = strconv.FormatFloat(wlID, 'f', 0, 64)
					}
				}
			}

			// If we have wordlist ID, check word count
			if wordlistIDStr != "" {
				wordlistID, err := strconv.ParseInt(wordlistIDStr, 10, 64)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
					c.Abort()
					return
				}

				// Only check for free users
				if user.SubscriptionPlan == model.PlanFree {
					wordCount, err := subService.CountWordsInWordlist(c.Request.Context(), wordlistID)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check word count"})
						c.Abort()
						return
					}

					if wordCount >= model.FreeWordsPerList {
						c.JSON(http.StatusPaymentRequired, gin.H{
							"error": "Free plan limit reached: maximum 10 words per wordlist",
							"code":  "SUBSCRIPTION_LIMIT_REACHED",
							"limit": model.FreeWordsPerList,
						})
						c.Abort()
						return
					}
				}
			}
		} else {
			// Check other subscription limits
			if err := subService.CheckSubscriptionLimits(c.Request.Context(), user.ID, action); err != nil {
				c.JSON(http.StatusPaymentRequired, gin.H{
					"error": err.Error(),
					"code":  "SUBSCRIPTION_LIMIT_REACHED",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func CheckoutRedirect() gin.HandlerFunc {
	return func(c *gin.Context) {
		expoURI := c.Query("redirect_uri")

		if expoURI != "" {
			c.Redirect(http.StatusSeeOther, expoURI)
		} else {
			c.Status(http.StatusBadRequest)
		}
	}
}
