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
	"github.com/stripe/stripe-go/v82/webhook"
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

// HandleStripeWebhook handles Stripe webhook events asynchronously
func HandleStripeWebhook(subService *service.SubscriptionService, jobService service.JobService) gin.HandlerFunc {
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

		// Verify webhook signature and construct event
		webhookSecret := subService.GetWebhookSecret()
		event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook signature verification failed"})
			return
		}

		// Enqueue the event for async processing
		_, err = jobService.ScheduleStripeWebhookJob(c.Request.Context(), event.ID, string(event.Type), event.Data.Raw)
		if err != nil {
			common.Logger.Error("Failed to enqueue Stripe webhook", "error", err, "event_id", event.ID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
			return
		}

		common.Logger.Info("Stripe webhook enqueued successfully", "event_id", event.ID, "event_type", event.Type)
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
func CheckSubscriptionLimits(subService *service.SubscriptionService, revenueCatService service.RevenueCatService, action string) gin.HandlerFunc {
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

		// Parse request body once to check for both optimistic flag and wordlist ID
		var hasOptimisticSubscription bool
		var wordlistIDFromBody string
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err == nil {
			// Check for optimistic subscription flag
			if flag, ok := body["hasOptimisticSubscription"].(bool); ok {
				hasOptimisticSubscription = flag
			}
			// Check for wordlist ID in body (for add_word operations)
			if wlID, ok := body["wordlistId"].(float64); ok {
				wordlistIDFromBody = strconv.FormatFloat(wlID, 'f', 0, 64)
			}
		}

		// If optimistic flag is present, verify subscription with RevenueCat
		if hasOptimisticSubscription && user.SubscriptionPlan == model.PlanFree {
			// Verify with RevenueCat to check if user actually has premium subscription
			hasPremium, err := revenueCatService.HasPremiumSubscription(c.Request.Context(), strconv.FormatInt(user.ID, 10))
			if err != nil {
				common.Logger.Error("Failed to verify subscription with RevenueCat", "error", err, "user_id", user.ID)
				// Continue with local subscription check - don't fail on RevenueCat API errors
			} else if hasPremium {
				// User has premium subscription in RevenueCat but not locally - allow the operation
				// The webhook will eventually sync the subscription status
				common.Logger.Info("Allowing operation for user with RevenueCat premium subscription", "user_id", user.ID)
				c.Next()
				return
			}
		}

		// For word operations, we might need wordlist ID
		if action == "add_word" {
			// Get wordlist ID from path, query, or body (already parsed above)
			wordlistIDStr := c.Param("wordlistId")
			if wordlistIDStr == "" {
				wordlistIDStr = c.Query("wordlistId")
			}
			if wordlistIDStr == "" {
				wordlistIDStr = wordlistIDFromBody
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
