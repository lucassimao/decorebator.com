package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"decorebator.com/internal/app"
	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

const mobileIAPRequestLimit = 16 << 10

type mobileIAPVerifyInput struct {
	TransactionID string `json:"transactionId"`
	PurchaseToken string `json:"purchaseToken"`
}

type mobileIAPRestoreInput struct {
	Store    model.EntitlementStore `json:"store"`
	Evidence []string               `json:"evidence"`
}

func RegisterMobileIAPRoutes(
	router *gin.Engine,
	appCtx *app.Context,
	limiter *service.StoreIAPRequestLimiter,
	authenticate gin.HandlerFunc,
) {
	routes := router.Group("/subscription/iap")
	routes.Use(authenticate, ResolveEffectiveSubscription(appCtx.EffectiveAccessService), SentryUserContextMiddleware())
	routes.Use(TimeoutMiddleware(25 * time.Second))
	routes.GET("/context", getMobileIAPContext(appCtx))
	routes.POST("/apple/verify", verifyMobileIAPApple(appCtx, limiter))
	routes.POST("/google/verify", verifyMobileIAPGoogle(appCtx, limiter))
	routes.POST("/restore", restoreMobileIAP(appCtx, limiter))
}

func getMobileIAPContext(appCtx *app.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !mobileIAPAvailable(c, appCtx) {
			return
		}
		store := model.EntitlementStore(c.Query("store"))
		if !store.Valid() {
			writeMobileIAPResponse(c, http.StatusBadRequest, invalidMobileIAPHTTPResponse())
			return
		}
		response, err := appCtx.StoreIAPOrchestrator.Context(c.Request.Context(), c.GetInt64("userID"), store)
		if err != nil {
			common.Logger.WarnContext(c.Request.Context(), "mobile IAP context unavailable", "store", store)
			writeMobileIAPResponse(c, http.StatusServiceUnavailable, unavailableMobileIAPHTTPResponse())
			return
		}
		writeMobileIAPResponse(c, http.StatusOK, response)
	}
}

func verifyMobileIAPApple(appCtx *app.Context, limiter *service.StoreIAPRequestLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !mobileIAPAvailable(c, appCtx) {
			return
		}
		var input mobileIAPVerifyInput
		if !decodeMobileIAPBody(c, &input) || strings.TrimSpace(input.TransactionID) == "" || input.PurchaseToken != "" {
			writeMobileIAPResponse(c, http.StatusBadRequest, invalidMobileIAPHTTPResponse())
			return
		}
		release, allowed := acquireMobileIAPLimit(c, limiter, "verify")
		if !allowed {
			return
		}
		defer release()
		response, err := appCtx.StoreIAPOrchestrator.VerifyApple(
			c.Request.Context(), c.GetInt64("userID"), input.TransactionID,
		)
		writeMobileIAPOperation(c, response, err)
	}
}

func verifyMobileIAPGoogle(appCtx *app.Context, limiter *service.StoreIAPRequestLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !mobileIAPAvailable(c, appCtx) {
			return
		}
		var input mobileIAPVerifyInput
		if !decodeMobileIAPBody(c, &input) || strings.TrimSpace(input.PurchaseToken) == "" || input.TransactionID != "" {
			writeMobileIAPResponse(c, http.StatusBadRequest, invalidMobileIAPHTTPResponse())
			return
		}
		release, allowed := acquireMobileIAPLimit(c, limiter, "verify")
		if !allowed {
			return
		}
		defer release()
		response, err := appCtx.StoreIAPOrchestrator.VerifyGoogle(
			c.Request.Context(), c.GetInt64("userID"), input.PurchaseToken,
		)
		writeMobileIAPOperation(c, response, err)
	}
}

func restoreMobileIAP(appCtx *app.Context, limiter *service.StoreIAPRequestLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !mobileIAPAvailable(c, appCtx) {
			return
		}
		var input mobileIAPRestoreInput
		if !decodeMobileIAPBody(c, &input) || !input.Store.Valid() || input.Evidence == nil {
			writeMobileIAPResponse(c, http.StatusBadRequest, invalidMobileIAPHTTPResponse())
			return
		}
		if len(input.Evidence) > service.MobileIAPRestoreEvidenceLimit {
			writeMobileIAPResponse(c, http.StatusRequestEntityTooLarge, invalidMobileIAPHTTPResponse())
			return
		}
		release, allowed := acquireMobileIAPLimit(c, limiter, "restore")
		if !allowed {
			return
		}
		defer release()
		response, err := appCtx.StoreIAPOrchestrator.Restore(
			c.Request.Context(), c.GetInt64("userID"), input.Store, input.Evidence,
		)
		writeMobileIAPOperation(c, response, err)
	}
}

func mobileIAPAvailable(c *gin.Context, appCtx *app.Context) bool {
	if appCtx.StoreIAPEnabled && appCtx.StoreIAPOrchestrator != nil {
		return true
	}
	writeMobileIAPResponse(c, http.StatusServiceUnavailable, unavailableMobileIAPHTTPResponse())
	return false
}

func decodeMobileIAPBody(c *gin.Context, target any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, mobileIAPRequestLimit)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return false
	}
	return decoder.Decode(&struct{}{}) == io.EOF
}

func acquireMobileIAPLimit(c *gin.Context, limiter *service.StoreIAPRequestLimiter, operation string) (func(), bool) {
	if limiter == nil {
		writeMobileIAPResponse(c, http.StatusServiceUnavailable, unavailableMobileIAPHTTPResponse())
		return nil, false
	}
	release, allowed, err := limiter.Acquire(c.Request.Context(), c.GetInt64("userID"), operation)
	if err != nil {
		common.Logger.WarnContext(c.Request.Context(), "mobile IAP limiter unavailable", "operation", operation)
		writeMobileIAPResponse(c, http.StatusServiceUnavailable, unavailableMobileIAPHTTPResponse())
		return nil, false
	}
	if !allowed {
		writeMobileIAPResponse(c, http.StatusTooManyRequests, unavailableMobileIAPHTTPResponse())
		return nil, false
	}
	return release, true
}

func writeMobileIAPOperation(c *gin.Context, response model.MobileIAPResponse, err error) {
	if err != nil {
		common.Logger.WarnContext(c.Request.Context(), "mobile IAP operation unavailable")
		writeMobileIAPResponse(c, http.StatusServiceUnavailable, unavailableMobileIAPHTTPResponse())
		return
	}
	status := http.StatusOK
	if response.Error != nil {
		if response.Error.Retryable {
			status = http.StatusServiceUnavailable
		} else {
			status = http.StatusBadRequest
		}
	}
	writeMobileIAPResponse(c, status, response)
}

func writeMobileIAPResponse(c *gin.Context, status int, response model.MobileIAPResponse) {
	if err := response.Validate(); err != nil {
		common.Logger.ErrorContext(c.Request.Context(), "invalid mobile IAP response contract")
		c.JSON(http.StatusInternalServerError, unavailableMobileIAPHTTPResponse())
		return
	}
	c.JSON(status, response)
}

func unavailableMobileIAPHTTPResponse() model.MobileIAPResponse {
	return model.MobileIAPResponse{
		Products: []model.MobileIAPProduct{},
		Error: &model.MobileIAPError{
			Code:       model.EntitlementResultRetryableProvider,
			MessageKey: "iap.error.unavailable", Retryable: true,
		},
		ServerTime: time.Now().UTC(),
	}
}

func invalidMobileIAPHTTPResponse() model.MobileIAPResponse {
	return model.MobileIAPResponse{
		Products: []model.MobileIAPProduct{},
		Error: &model.MobileIAPError{
			Code:       model.EntitlementResultInvalidPurchase,
			MessageKey: "iap.error.invalid_purchase",
		},
		ServerTime: time.Now().UTC(),
	}
}
