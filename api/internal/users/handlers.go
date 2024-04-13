package users

import (
	"errors"
	"fmt"
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
			c.Writer.Header().Set("Authorization", jwtToken)
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
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"token": jwtToken})
	}
}

func (h *UserHandlers) Authenticate(c *gin.Context) {

	const BearerSchema = "Bearer "
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, BearerSchema)

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
