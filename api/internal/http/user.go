package http

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"decorebator.com/internal/common"
	"decorebator.com/internal/config"
	"decorebator.com/internal/mail"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
)

type SignupInput struct {
	FirstName         string `json:"firstName" binding:"required"`
	LastName          string `json:"lastName" binding:"required"`
	Email             string `json:"email" binding:"required"`
	Password          string `json:"password" binding:"required"`
	Country           string `json:"country"`           // Optional ISO 3166-1 alpha-2 country code
	PreferredLanguage string `json:"preferredLanguage"` // Optional language code
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ResetPasswordInput struct {
	Password string `json:"password" binding:"required"`
	Token    string `json:"token" binding:"required"`
}

type RequestResetPasswordEmailInput struct {
	Email string `json:"email" binding:"required"`
}

type UpdateProfileInput struct {
	FirstName                 *string                    `json:"firstName"`
	LastName                  *string                    `json:"lastName"`
	Country                   *string                    `json:"country"`
	DateOfBirth               *string                    `json:"dateOfBirth"` // Expect ISO format: YYYY-MM-DD
	PreferredLanguage         *string                    `json:"preferredLanguage"`
	NotificationsEnabled      *bool                      `json:"notificationsEnabled"`
	UpdateProfilePictureInput *UpdateProfilePictureInput `json:"updateProfilePicture"`
	UpdatePasswordInput       *UpdatePasswordInput       `json:"updatePassword"`
}

type UpdatePasswordInput struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type UpdateProfilePictureInput struct {
	Base64Data string `json:"base64Data"`
	Extension  string `json:"extension,omitempty"`
}

const updateProfileRequestLimit = 8 << 20
const profileCleanupTimeout = 1500 * time.Millisecond
const profileReconciliationScheduleTimeout = 500 * time.Millisecond

type profileUploadAdmission struct {
	mu     sync.Mutex
	users  map[int64]struct{}
	global chan struct{}
}

func newProfileUploadAdmission(maxConcurrent int) *profileUploadAdmission {
	return &profileUploadAdmission{users: make(map[int64]struct{}), global: make(chan struct{}, maxConcurrent)}
}

func (a *profileUploadAdmission) tryAcquire(userID int64) (func(), bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, exists := a.users[userID]; exists {
		return nil, false
	}
	select {
	case a.global <- struct{}{}:
		a.users[userID] = struct{}{}
	default:
		return nil, false
	}
	return func() {
		a.mu.Lock()
		delete(a.users, userID)
		<-a.global
		a.mu.Unlock()
	}, true
}

type UserRoutes struct {
	userService           *service.UserService
	authSessions          userAuthSessions
	authLimiter           *service.AuthRateLimiter
	jobService            service.JobService
	httpSecurity          config.HTTPSecurityConfig
	profileBodies         chan struct{}
	profileUploads        *profileUploadAdmission
	profileUserExists     func(context.Context, int64) (bool, error)
	normalizeProfileImage func(context.Context, string) (common.ProfileImage, error)
}

type userAuthSessions interface {
	Logout(context.Context, string) error
	Refresh(context.Context, string) (service.SessionCredentials, error)
}

// NewUserRoutes creates a new UserRoutes with injected dependencies
func NewUserRoutes(
	userService *service.UserService,
	authSessions *service.AuthSessionService,
	authLimiter *service.AuthRateLimiter,
	jobService service.JobService,
	httpSecurity config.HTTPSecurityConfig,
) *UserRoutes {
	return &UserRoutes{
		userService:           userService,
		authSessions:          authSessions,
		authLimiter:           authLimiter,
		jobService:            jobService,
		httpSecurity:          httpSecurity,
		profileBodies:         make(chan struct{}, 8),
		profileUploads:        newProfileUploadAdmission(2),
		profileUserExists:     userService.Exists,
		normalizeProfileImage: common.NormalizeProfileImage,
	}
}

func translateValidationErrors(errs validator.ValidationErrors) map[string]string {
	var errors = make(map[string]string)
	for _, e := range errs {
		field := strings.ToLower(strings.SplitN(e.StructNamespace(), ".", 2)[1])
		switch e.Tag() {
		case "required":
			errors[field] = "The " + field + " field is required."
		case "email":
			errors[field] = "The " + field + " field must be a valid email address."
		case "min":
			errors[field] = "The " + field + " field must be at least " + e.Param() + " characters long."
		default:
			errors[field] = "The " + field + " field is invalid."
		}
	}
	return errors
}

func (h *UserRoutes) SignUp(c *gin.Context) {
	txn := sentry.StartTransaction(c.Request.Context(), "auth.signup", sentry.WithDescription("UserRoutes.SignUp"))
	defer txn.Finish()
	c.Request = c.Request.WithContext(txn.Context())

	var input SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		var ve validator.ValidationErrors
		var body any

		if errors.As(err, &ve) {
			body = gin.H{"validationErrors": translateValidationErrors(ve)}
		} else {
			body = gin.H{"error": err.Error()}
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, body)
		return
	}
	if err := common.ValidatePassword(input.Password); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"validationErrors": gin.H{"password": err.Error()},
		})
		return
	}

	canonicalEmail, err := common.NormalizeEmail(input.Email)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"validationErrors": gin.H{"email": "The email field must be a valid email address."},
		})
		return
	}
	input.Email = canonicalEmail
	if !applyAuthLimit(c, h.authLimiter, service.AuthLimitSignup, service.AuthLimitAccount, canonicalEmail) {
		return
	}

	// Prepare country parameter (convert empty string to nil)
	var country *string
	if input.Country != "" {
		country = &input.Country
	}

	// Prepare preferredLanguage parameter (convert empty string to nil)
	var preferredLanguage *string
	if input.PreferredLanguage != "" {
		preferredLanguage = &input.PreferredLanguage
	}

	span := sentry.StartSpan(c.Request.Context(), "auth.user.create", sentry.WithDescription("userService.RegisterPendingUser"))
	err = h.userService.RegisterPendingUser(
		span.Context(), input.FirstName, input.LastName, input.Password, input.Email, country, preferredLanguage,
	)
	span.Finish()

	if err != nil {
		common.Logger.ErrorContext(c.Request.Context(), "failed to register account", "error", err)
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Account registration temporarily unavailable"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "If the address can be used, account instructions will be sent."})
}

func (h *UserRoutes) Login(c *gin.Context) {
	txn := sentry.StartTransaction(c.Request.Context(), "auth.login", sentry.WithDescription("UserRoutes.Login"))
	defer txn.Finish()
	c.Request = c.Request.WithContext(txn.Context())

	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	canonicalEmail, err := common.NormalizeEmail(input.Email)
	if err != nil {
		canonicalEmail = "invalid:" + strings.ToLower(input.Email)
	}
	if !applyAuthLimit(c, h.authLimiter, service.AuthLimitLogin, service.AuthLimitAccount, canonicalEmail) {
		return
	}
	span := sentry.StartSpan(c.Request.Context(), "auth.user.login", sentry.WithDescription("userService.LoginUser"))
	credentials, err := h.userService.LoginUser(span.Context(), input.Email, input.Password)
	span.Finish()
	if err != nil {
		if errors.Is(err, service.ErrInvalidLoginCredentials) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password"})
			return
		}
		if releaseErr := h.authLimiter.Release(
			c.Request.Context(), service.AuthLimitLogin, service.AuthLimitAccount, canonicalEmail,
		); releaseErr != nil {
			common.Logger.ErrorContext(c.Request.Context(), "failed to refund login limiter reservation", "error", releaseErr)
		}
		common.Logger.ErrorContext(c.Request.Context(), "login infrastructure failure", "error", err)
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication temporarily unavailable"})
		return
	}
	if clearErr := h.authLimiter.Clear(
		c.Request.Context(), service.AuthLimitLogin, service.AuthLimitAccount, canonicalEmail,
	); clearErr != nil {
		common.Logger.ErrorContext(c.Request.Context(), "failed to clear successful login limiter bucket", "error", clearErr)
	}

	writeSessionCredentials(c, h.httpSecurity, credentials)
	c.Status(http.StatusOK)
}

func (h *UserRoutes) Logout(c *gin.Context) {
	// The browser must lose both HttpOnly capabilities even when server-side
	// revocation is temporarily unavailable. The client keeps a durable retry
	// barrier for the uncertain revocation outcome.
	writeAuthenticationCookie(c, h.httpSecurity, "")
	writeRefreshCookie(c, h.httpSecurity, "")
	if err := h.authSessions.Logout(c.Request.Context(), readRefreshToken(c, h.httpSecurity)); err != nil {
		common.Logger.ErrorContext(c.Request.Context(), "failed to revoke auth session", "error", err)
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Logout temporarily unavailable"})
		return
	}
	c.Status(http.StatusOK)
}

func (h *UserRoutes) Refresh(c *gin.Context) {
	credentials, err := h.authSessions.Refresh(c.Request.Context(), readRefreshToken(c, h.httpSecurity))
	if err != nil {
		if errors.Is(err, repository.ErrInvalidRefreshToken) ||
			errors.Is(err, repository.ErrRefreshTokenReplay) ||
			errors.Is(err, repository.ErrSessionExpired) {
			writeAuthenticationCookie(c, h.httpSecurity, "")
			writeRefreshCookie(c, h.httpSecurity, "")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
		} else {
			common.Logger.ErrorContext(c.Request.Context(), "failed to refresh auth session", "error", err)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication temporarily unavailable"})
		}
		return
	}
	writeSessionCredentials(c, h.httpSecurity, credentials)
	c.Status(http.StatusOK)
}

func (h *UserRoutes) ResetPassword(c *gin.Context) {
	var input ResetPasswordInput

	if err := c.ShouldBindJSON(&input); err != nil {
		var ve validator.ValidationErrors
		var body any

		if errors.As(err, &ve) {
			body = gin.H{"validationErrors": translateValidationErrors(ve)}
		} else {
			body = gin.H{"error": err.Error()}
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, body)
		return
	}

	payload, err := mail.ValidateResetPasswordPayload(input.Token)
	if err != nil {
		if err.Error() == "token expired" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token_expired"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "token_invalid"})
		return
	}
	var resetErr error
	if payload.Legacy {
		resetErr = h.userService.ResetLegacyPassword(
			c.Request.Context(), payload.UserID, input.Token, input.Password,
		)
	} else {
		resetErr = h.userService.ResetPasswordAndVerifyEmail(
			c.Request.Context(), payload.UserID, payload.TokenID, input.Password,
		)
	}
	if resetErr != nil {
		if errors.Is(resetErr, service.ErrAuthHardeningWritesDisabled) {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Password reset temporarily unavailable"})
			return
		}
		var businessError common.BusinessError
		if errors.As(resetErr, &businessError) {
			if businessError.Error() == "token_invalid" && !applyAuthLimit(
				c, h.authLimiter, service.AuthLimitResetConsume, service.AuthLimitAccount,
				fmt.Sprintf("user:%d", payload.UserID),
			) {
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": businessError.Error()})
			return
		}
		common.Logger.ErrorContext(c.Request.Context(), "failed to update password", "userId", payload.UserID, "error", resetErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}
	c.Status(http.StatusOK)
}

func (h *UserRoutes) SendResetPasswordEmail(c *gin.Context) {
	var input RequestResetPasswordEmailInput

	if err := c.ShouldBindJSON(&input); err != nil {
		var ve validator.ValidationErrors
		var body any

		if errors.As(err, &ve) {
			body = gin.H{"validationErrors": translateValidationErrors(ve)}
		} else {
			body = gin.H{"error": err.Error()}
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, body)
		return
	}
	canonicalEmail, err := common.NormalizeEmail(input.Email)
	queuedEmail := canonicalEmail
	if err != nil {
		digest := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(input.Email))))
		canonicalEmail = fmt.Sprintf("invalid:%x", digest)
		queuedEmail = ""
	}
	if !applyAuthLimit(
		c, h.authLimiter, service.AuthLimitResetRequest, service.AuthLimitAccount, canonicalEmail,
	) {
		return
	}

	if err := h.jobService.ScheduleResetPasswordEmailJob(c.Request.Context(), queuedEmail); err != nil {
		common.Logger.ErrorContext(c.Request.Context(), "failed to enqueue reset password email", "error", err)
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Password reset temporarily unavailable"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "If the account exists, reset instructions will be sent."})
}

func writeAuthenticationCookie(c *gin.Context, policy config.HTTPSecurityConfig, jwtToken string) {
	maxAge := int64(config.AccessTokenDuration / time.Second)

	if !canConvertToInt(maxAge) {
		panic("maxAge can not be safely converted to int")
	}

	c.SetSameSite(http.SameSiteStrictMode)
	cookieName := authenticationCookieName(policy)

	if len(jwtToken) > 0 {
		c.SetCookie(cookieName, jwtToken, int(maxAge), "/", policy.CookieDomain, policy.SecureCookies, true)
	} else {
		// clear cookie
		common.Logger.DebugContext(c.Request.Context(), "Clearing authorization cookie")
		c.SetCookie(cookieName, "", int(-1), "/", policy.CookieDomain, policy.SecureCookies, true)
	}
}

func writeSessionCredentials(c *gin.Context, policy config.HTTPSecurityConfig, credentials service.SessionCredentials) {
	if isNativeCredentialRequest(c) {
		c.Header("Authorization", credentials.AccessToken)
		c.Header("X-Refresh-Token", credentials.RefreshToken)
		return
	}
	writeAuthenticationCookie(c, policy, credentials.AccessToken)
	writeRefreshCookie(c, policy, credentials.RefreshToken)
}

func writeRefreshCookie(c *gin.Context, policy config.HTTPSecurityConfig, refreshToken string) {
	maxAge := int(service.RefreshIdleDuration / time.Second)
	if refreshToken == "" {
		maxAge = -1
	}
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshCookieName(policy), refreshToken, maxAge, "/", policy.CookieDomain, policy.SecureCookies, true)
}

func readRefreshToken(c *gin.Context, policy config.HTTPSecurityConfig) string {
	if isNativeCredentialRequest(c) {
		return c.GetHeader("X-Refresh-Token")
	}
	if token, err := c.Cookie(refreshCookieName(policy)); err == nil {
		return token
	}
	return ""
}

func authenticationCookieName(policy config.HTTPSecurityConfig) string {
	if policy.SecureCookies {
		return "__Host-Authorization"
	}
	return "Authorization"
}

func refreshCookieName(policy config.HTTPSecurityConfig) string {
	if policy.SecureCookies {
		return "__Host-RefreshToken"
	}
	return "RefreshToken"
}

func isNativeCredentialRequest(c *gin.Context) bool {
	return c.GetHeader("X-Auth-Client") == "native" && len(c.Request.Header.Values("Origin")) == 0
}

func canConvertToInt(n int64) bool {
	return n >= int64(math.MinInt) && n <= int64(math.MaxInt)
}

func (h *UserRoutes) UpdateProfile(c *gin.Context) { //nolint:gocyclo // staged profile mutation keeps every side effect explicit
	// Get user from context (set by auth middleware)
	userAsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	user, ok := userAsAny.(*model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user"})
		return
	}
	if c.Request.ContentLength > updateProfileRequestLimit {
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Profile update is too large."})
		return
	}
	select {
	case h.profileBodies <- struct{}{}:
	default:
		c.Header("Retry-After", "1")
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Profile updates are temporarily busy."})
		return
	}
	bodyAdmissionHeld := true
	defer func() {
		if bodyAdmissionHeld {
			<-h.profileBodies
		}
	}()
	clearReadDeadline := setProfileReadDeadline(c.Request.Context(), c.Writer)
	defer clearReadDeadline()
	stopBodyCancellation := closeBodyOnContextDone(c.Request.Context(), c.Request.Body)
	defer stopBodyCancellation()

	var input UpdateProfileInput
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, updateProfileRequestLimit)
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&input); err != nil {
		clearReadDeadline()
		writeProfileDecodeError(c, err)
		return
	}
	if err := requireJSONWhitespaceEOF(decoder, c.Request.Body); err != nil {
		clearReadDeadline()
		writeProfileDecodeError(c, err)
		return
	}
	clearReadDeadline()
	var releaseUpload func()
	if input.UpdateProfilePictureInput != nil {
		var acquired bool
		releaseUpload, acquired = h.profileUploads.tryAcquire(user.ID)
		if !acquired {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Another profile picture update is in progress."})
			return
		}
		defer releaseUpload()
	}
	if err := validateProfileInputBounds(input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	<-h.profileBodies
	bodyAdmissionHeld = false

	var newPassword *string
	if input.UpdatePasswordInput != nil {
		err := h.userService.VerifyPassword(c.Request.Context(), user.Email, input.UpdatePasswordInput.CurrentPassword)
		if err != nil {
			if c.Request.Context().Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Profile update timed out."})
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Could not update password."})
			return
		}

		newPassword = &input.UpdatePasswordInput.NewPassword
	}
	if err := service.ValidateProfileUpdate(input.FirstName, input.LastName, newPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse and validate every non-image field before creating an object.
	var dateOfBirth *time.Time
	if input.DateOfBirth != nil && *input.DateOfBirth != "" {
		parsedDate, err := time.Parse("2006-01-02", *input.DateOfBirth)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Expected: YYYY-MM-DD"})
			return
		}
		dateOfBirth = &parsedDate
	}

	// Upload profile picture if provided
	var url *string
	var uploadedReceipt common.UploadReceipt
	var cleanupIntentID int64
	var err error
	if input.UpdateProfilePictureInput != nil {
		exists, existenceErr := h.profileUserExists(c.Request.Context(), user.ID)
		if existenceErr != nil {
			if c.Request.Context().Err() != nil || errors.Is(existenceErr, context.Canceled) || errors.Is(existenceErr, context.DeadlineExceeded) {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Profile update timed out."})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user."})
			return
		}
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		cmd := input.UpdateProfilePictureInput

		image, imageErr := h.normalizeProfileImage(c.Request.Context(), cmd.Base64Data)
		if imageErr != nil {
			if errors.Is(imageErr, common.ErrProfileImageBusy) {
				c.Header("Retry-After", "1")
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Profile image processing is busy."})
				return
			}
			if errors.Is(imageErr, context.Canceled) || errors.Is(imageErr, context.DeadlineExceeded) {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Profile update timed out."})
				return
			}
			if errors.Is(imageErr, common.ErrProfileImageTooLarge) {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Profile image is too large."})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image format."})
			return
		}
		legacyExtension := strings.ToLower(strings.TrimSpace(cmd.Extension))
		if legacyExtension != "" && legacyExtension != image.Extension && !(legacyExtension == "jpeg" && image.Extension == "jpg") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image format."})
			return
		}

		objectNonce := common.GenerateRandomString(16)
		if objectNonce == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload profile picture."})
			return
		}
		objectName := fmt.Sprintf("users/%d-%s.%s", user.ID, objectNonce, image.Extension)
		profileBucket := common.MinIOBucketName()
		plannedURL, urlErr := common.MinIOObjectURL(profileBucket, objectName)
		if urlErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare profile picture upload."})
			return
		}
		// Commit a reference-aware exact-key cleanup intent before the external
		// side effect. A crash or database outage after PUT cannot orphan the key.
		cleanupIntentID, err = h.userService.ScheduleUncertainProfileUploadCleanup(
			c.Request.Context(), user.ID, profileBucket, objectName, plannedURL,
		)
		if err != nil {
			if c.Request.Context().Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Profile update timed out."})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Profile picture upload is temporarily unavailable."})
			}
			return
		}
		uploadResult, uploadErr := common.UploadWithReceipt(c.Request.Context(), image.Data, profileBucket, objectName, image.ContentType)
		if uploadErr != nil {
			// The random key cannot collide with another upload. Remove every
			// version in case the object store committed before an ambiguous error.
			cleanupErr := runProfileCleanup(c.Request.Context(), func(cleanupContext context.Context) error {
				return common.MinIODeleteObjectAllVersions(cleanupContext, profileBucket, objectName)
			})
			if cleanupErr != nil {
				common.Logger.ErrorContext(context.WithoutCancel(c.Request.Context()), "failed immediate uncertain profile picture cleanup", "object", objectName, "error", cleanupErr)
			}
			if errors.Is(uploadErr, context.Canceled) || errors.Is(uploadErr, context.DeadlineExceeded) || c.Request.Context().Err() != nil {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Profile update timed out."})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload profile picture."})
			}
			return
		}
		uploadedReceipt = uploadResult
		url = &uploadResult.URL
	}

	// Update user profile
	var updatedUser *service.User
	if uploadedReceipt.ObjectName != "" {
		updatedUser, err = h.userService.UpdateProfileWithUploadIntent(c.Request.Context(), user.ID, input.FirstName, input.LastName, input.Country, input.PreferredLanguage, url, newPassword, dateOfBirth, input.NotificationsEnabled, cleanupIntentID, uploadedReceipt)
	} else {
		updatedUser, err = h.userService.UpdateProfile(c.Request.Context(), user.ID, input.FirstName, input.LastName, input.Country, input.PreferredLanguage, url, newPassword, dateOfBirth, input.NotificationsEnabled)
	}
	persistenceAmbiguous := err != nil && !isDefinitiveProfilePersistenceFailure(err)
	if err != nil {
		if uploadedReceipt.ObjectName != "" {
			var cleanupErr error
			if !persistenceAmbiguous {
				cleanupErr = runProfileCleanup(c.Request.Context(), func(cleanupContext context.Context) error {
					return h.userService.CompensateProfileUpload(cleanupContext, user.ID, uploadedReceipt)
				})
			}
			if cleanupErr != nil && c.Request.Context().Err() == nil {
				reconcileContext, cancelReconcile := context.WithTimeout(c.Request.Context(), profileReconciliationScheduleTimeout)
				reconcileErr := h.userService.ScheduleProfileUploadReconciliation(reconcileContext, user.ID, uploadedReceipt)
				cancelReconcile()
				if reconcileErr != nil {
					common.Logger.ErrorContext(context.WithoutCancel(c.Request.Context()), "failed to compensate profile picture upload", "object", uploadedReceipt.ObjectName, "error", errors.Join(cleanupErr, reconcileErr))
				}
			}
		}
		if profileRequestTimedOut(c.Request.Context(), err) {
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "Profile update timed out."})
			return
		}
		switch err.(type) {
		case common.BusinessError:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		}
		return
	}

	// Remove password hash from response
	updatedUser.PasswordHash = ""
	if newPassword != nil {
		writeAuthenticationCookie(c, h.httpSecurity, "")
		writeRefreshCookie(c, h.httpSecurity, "")
	}

	c.JSON(http.StatusOK, updatedUser)
}

func setProfileReadDeadline(ctx context.Context, writer http.ResponseWriter) func() {
	deadline, ok := ctx.Deadline()
	if !ok {
		return func() {}
	}
	controller := http.NewResponseController(writer)
	if err := controller.SetReadDeadline(deadline); err != nil {
		return func() {}
	}
	return func() { _ = controller.SetReadDeadline(time.Time{}) }
}

func runProfileCleanup(ctx context.Context, cleanup func(context.Context) error) error {
	cleanupContext, cancelCleanup := context.WithTimeout(ctx, profileCleanupTimeout)
	defer cancelCleanup()
	return cleanup(cleanupContext)
}

func profileRequestTimedOut(ctx context.Context, err error) bool {
	return ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func isDefinitiveProfilePersistenceFailure(err error) bool {
	var businessError common.BusinessError
	if errors.As(err, &businessError) {
		return true
	}
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError)
}

func validateProfileInputBounds(input UpdateProfileInput) error {
	if input.FirstName != nil && (len(*input.FirstName) > 400 || utf8.RuneCountInString(*input.FirstName) > 100) {
		return errors.New("first name is too long")
	}
	if input.LastName != nil && (len(*input.LastName) > 400 || utf8.RuneCountInString(*input.LastName) > 100) {
		return errors.New("last name is too long")
	}
	if input.Country != nil && len(*input.Country) > 2 {
		return errors.New("country code is too long")
	}
	if input.PreferredLanguage != nil && len(*input.PreferredLanguage) > 10 {
		return errors.New("preferred language is too long")
	}
	if input.DateOfBirth != nil && len(*input.DateOfBirth) > len("2006-01-02") {
		return errors.New("date of birth is invalid")
	}
	if input.UpdatePasswordInput != nil && len(input.UpdatePasswordInput.CurrentPassword) > common.PasswordMaxBytes {
		return errors.New("current password is too long")
	}
	if input.UpdatePasswordInput != nil && len(input.UpdatePasswordInput.NewPassword) > common.PasswordMaxBytes {
		return errors.New("new password is too long")
	}
	return nil
}

func requireJSONWhitespaceEOF(decoder *json.Decoder, body io.Reader) error {
	reader := io.MultiReader(decoder.Buffered(), body)
	buffer := make([]byte, 4096)
	hasTrailingValue := false
	for {
		count, err := reader.Read(buffer)
		for _, value := range buffer[:count] {
			if value != ' ' && value != '\t' && value != '\r' && value != '\n' {
				hasTrailingValue = true
			}
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}
	}
	if hasTrailingValue {
		return errors.New("request body must contain one JSON value")
	}
	return nil
}

func writeProfileDecodeError(c *gin.Context, err error) {
	if c.Request.Context().Err() != nil {
		closeProfileConnection(c)
		c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "Profile update timed out."})
		return
	}
	var timeoutError interface{ Timeout() bool }
	if errors.As(err, &timeoutError) && timeoutError.Timeout() {
		closeProfileConnection(c)
		c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "Profile update timed out."})
		return
	}
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Profile update is too large."})
		return
	}
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"validationErrors": translateValidationErrors(validationErrors)})
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid profile update."})
}

func closeProfileConnection(c *gin.Context) {
	c.Header("Connection", "close")
	c.Request.Close = true
}

func closeBodyOnContextDone(ctx context.Context, body io.Closer) func() {
	done := make(chan struct{})
	var once sync.Once
	go func() {
		select {
		case <-ctx.Done():
			_ = body.Close()
		case <-done:
		}
	}()
	return func() { once.Do(func() { close(done) }) }
}

func (h *UserRoutes) GetProfile(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userIDAny, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found."})
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type in context"})
		return
	}

	user, _, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		// If user not found (e.g., deleted), return 401 instead of 500
		if err.Error() == "user not found" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error."})
		}
		return
	}

	// hide unnecessary data
	user.PasswordHash = ""
	user.StripeCustomerID = nil

	c.JSON(http.StatusOK, user)
}

func (h *UserRoutes) DeleteProfile(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userIDAny, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found."})
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type in context"})
		return
	}

	err := h.userService.Delete(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User could not be deleted."})
		return
	}

	writeAuthenticationCookie(c, h.httpSecurity, "")
	writeRefreshCookie(c, h.httpSecurity, "")
	c.Status(http.StatusNoContent)
}
