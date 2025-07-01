package http

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/mail"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SignupInput struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=5"`
	Country   string `json:"country"` // Optional ISO 3166-1 alpha-2 country code
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
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
	UpdateProfilePictureInput *UpdateProfilePictureInput `json:"updateProfilePicture"`
	UpdatePasswordInput       *UpdatePasswordInput       `json:"updatePassword"`
}

type UpdatePasswordInput struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type UpdateProfilePictureInput struct {
	Base64Data string `json:"base64Data"`
	Extension  string `json:"extension"`
}

type UserRoutes struct{}

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

	// Prepare country parameter (convert empty string to nil)
	var country *string
	if input.Country != "" {
		country = &input.Country
	}

	user, err := service.SaveUser(input.FirstName, input.LastName, input.Password, input.Email, country)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		jwtToken, err := service.LoginUser(input.Email, input.Password)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("authorization", jwtToken)
		writeAuthenticationCookie(c, jwtToken)
		c.Status(http.StatusCreated)
		go mail.AddContactToList(user)
		go func() {
			if err := mail.SendWelcomeEmail(input.Email); err != nil {
				common.Logger.Error("failed to send welcome email", "email", input.Email, "error", err)
			}
		}()

	}
}

func (h *UserRoutes) Login(c *gin.Context) {
	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jwtToken, err := service.LoginUser(input.Email, input.Password)
	if err == nil {
		c.Header("authorization", jwtToken)
		writeAuthenticationCookie(c, jwtToken)
		c.Status(http.StatusOK)

	} else {
		c.Status(http.StatusBadRequest)
	}
}

func (h *UserRoutes) Logout(c *gin.Context) {
	writeAuthenticationCookie(c, "")
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
		c.Status(http.StatusBadRequest)
		return
	}

	if err := service.UpdatePassword(payload.UserId, input.Password); err != nil {
		common.Logger.Error("failed to update password", "userId", payload.UserId, "error", err)
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

	err := mail.SendResetPasswordEmail(input.Email)
	if err != nil {
		common.Logger.Error("failed to send reset password email", "error", err)
	}
	c.Status(http.StatusOK)
}

func writeAuthenticationCookie(c *gin.Context, jwtToken string) {
	var maxAge, path, domain, secure, httpOnly, sameSite = int64(0), "/", "localhost", false, true, http.SameSiteStrictMode

	if os.Getenv("ENV") == "production" {
		maxAge = service.AUTH_TOKEN_DURATION.Milliseconds()
		domain = "decorebator.com"
		// requires https
		secure = true
	}

	if !canConvertToInt(maxAge) {
		panic("maxAge can not be safely converted to int")
	}

	c.SetSameSite(sameSite)

	if len(jwtToken) > 0 {
		c.SetCookie("Authorization", jwtToken, int(maxAge), path, domain, secure, httpOnly)
	} else {
		// clear cookie
		common.Logger.Debug("Clearing authorization cookie")
		c.SetCookie("Authorization", "", int(-1), path, domain, secure, httpOnly)
	}
}

func canConvertToInt(n int64) bool {
	return n >= int64(math.MinInt) && n <= int64(math.MaxInt)
}

func (h *UserRoutes) UpdateProfile(c *gin.Context) {
	var input UpdateProfileInput

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

	var newPassword *string
	if input.UpdatePasswordInput != nil {
		_, err := service.LoginUser(user.Email, input.UpdatePasswordInput.CurrentPassword)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Could not update password."})
			return
		}

		newPassword = &input.UpdatePasswordInput.NewPassword
	}

	// Upload profile picture if provided
	var url *string
	if input.UpdateProfilePictureInput != nil {

		cmd := input.UpdateProfilePictureInput

		imgBytes, mimeType, err := common.DecodeImageBase64(cmd.Base64Data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image format."})
			return
		}

		objectName := fmt.Sprintf("users/%d-%d.%s", user.ID, time.Now().Unix(), cmd.Extension)
		uploadResult, err := common.Upload(imgBytes, "decorebator", objectName, mimeType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload profile picture."})
			return
		}
		url = &uploadResult
	}

	// Parse date of birth if provided
	var dateOfBirth *time.Time
	if input.DateOfBirth != nil && *input.DateOfBirth != "" {
		parsedDate, err := time.Parse("2006-01-02", *input.DateOfBirth)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Expected: YYYY-MM-DD"})
			return
		}
		dateOfBirth = &parsedDate
	}

	// Update user profile
	updatedUser, err := service.UpdateProfile(user.ID, input.FirstName, input.LastName, input.Country, input.PreferredLanguage, url, newPassword, dateOfBirth)
	if err != nil {
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

	c.JSON(http.StatusOK, updatedUser)
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

	user, err := service.GetProfile(userID)
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

	err := service.Delete(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User could not be deleted."})
		return
	}

	c.Status(http.StatusNoContent)
}
