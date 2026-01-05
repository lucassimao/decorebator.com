package http

import (
	"io"
	"net/http"
	"time"

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
			common.Logger.ErrorContext(c.Request.Context(), "Failed to enqueue Stripe webhook", "error", err, "event_id", event.ID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
			return
		}

		common.Logger.InfoContext(c.Request.Context(), "Stripe webhook enqueued successfully", "event_id", event.ID, "event_type", event.Type)
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

		// Get latest subscription for UI display (includes canceled subscriptions with remaining time)
		subscription, err := subRepo.GetSubscriptionForUser(c.Request.Context(), user.ID, false)
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
			// Add computed fields to help frontend determine subscription state
			response["isActive"] = subscription.IsActive()

			// Check if subscription is canceled but still has valid time remaining
			response["isCancelledButActive"] = subscription.Status == model.StatusCanceled &&
				subscription.CurrentPeriodEnd.After(time.Now())

			// Check if subscription is in grace period (past_due but still active)
			response["isInGracePeriod"] = subscription.Status == model.StatusPastDue &&
				subscription.IsActive()

			// Calculate days remaining for canceled subscriptions
			if subscription.Status == model.StatusCanceled && subscription.CurrentPeriodEnd.After(time.Now()) {
				daysRemaining := int(time.Until(subscription.CurrentPeriodEnd).Hours() / 24)
				response["daysRemaining"] = daysRemaining
			}
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
