package users

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SignupInput struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=5"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=5"`
}

type UserHandlers struct{}

var Handlers = &UserHandlers{}

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

func (h *UserHandlers) SignUp(c *gin.Context) {
	var input SignupInput
	if err := c.BindJSON(&input); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			c.JSON(http.StatusBadRequest, gin.H{"validationErrors": translateValidationErrors(ve)})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	saved, err := SaveUser(input.FirstName, input.LastName, input.Password, input.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		jwtToken, err := LoginUser(input.Email, input.Password)
		if err == nil {
			writeAuthenticationCookie(c, jwtToken)
		}
		c.JSON(http.StatusCreated, saved)
	}
}

func (h *UserHandlers) Login(c *gin.Context) {
	var input LoginInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jwtToken, err := LoginUser(input.Email, input.Password)
	if err != nil {
		c.Status(http.StatusBadRequest)
	} else {
		writeAuthenticationCookie(c, jwtToken)
		c.Status(http.StatusOK)
	}
}

func (h *UserHandlers) Logout(c *gin.Context) {
	writeAuthenticationCookie(c, "")
	c.Status(http.StatusOK)
}

func writeAuthenticationCookie(c *gin.Context, jwtToken string) {
	var maxAge, path, domain, secure, httpOnly, sameSite = int64(0), "/", "localhost", false, true, http.SameSiteStrictMode

	if os.Getenv("ENV") == "production" {
		maxAge = AUTH_TOKEN_DURATION.Milliseconds()
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
		fmt.Println("Clearing authorization cookie")
		c.SetCookie("Authorization", "", int(-1), path, domain, secure, httpOnly)
	}
}

func (h *UserHandlers) Authenticate(c *gin.Context) {

	const BearerSchema = "Bearer "
	authorization, err := c.Cookie("Authorization")

	if err == http.ErrNoCookie {
		// fallback to header
		authorization = c.GetHeader("Authorization")
	}

	if authorization == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization missing"})
		return
	}

	tokenString := strings.TrimPrefix(authorization, BearerSchema)

	if tokenString == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token not found"})
		return
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_KEY")), nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token vaidation error"})
		return
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)

	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.Set("userID", userID)

	c.Next()
}

func canConvertToInt(n int64) bool {
	return n >= int64(math.MinInt) && n <= int64(math.MaxInt)
}
