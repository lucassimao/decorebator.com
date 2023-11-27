package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

func (h *UserHandlers) SignUp(c *gin.Context) {
	var input SignupInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	saved, err := SaveUser(input.FirstName, input.LastName, input.Password, input.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
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
