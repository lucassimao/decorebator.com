package http

import (
	"errors"
	"math"
	"net/http"
	"os"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/mail"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SignupInput struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=5"`
}

type loginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
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
	if err := c.BindJSON(&input); err != nil {
		var ve validator.ValidationErrors
		var body any

		if errors.As(err, &ve) {
			body = gin.H{"validationErrors": translateValidationErrors(ve)}
		} else {
			body = gin.H{"error": err.Error()}
		}

		c.JSON(http.StatusBadRequest, body)
		return
	}

	user, err := service.SaveUser(input.FirstName, input.LastName, input.Password, input.Email)

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
	}
}

func (h *UserRoutes) Login(c *gin.Context) {
	var input loginInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jwtToken, err := service.LoginUser(input.Email, input.Password)
	if err != nil {
		c.Status(http.StatusBadRequest)
	} else {
		c.Header("authorization", jwtToken)
		writeAuthenticationCookie(c, jwtToken)
		c.Status(http.StatusOK)
	}
}

func (h *UserRoutes) Logout(c *gin.Context) {
	writeAuthenticationCookie(c, "")
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
