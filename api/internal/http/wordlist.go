package http

import (
	"errors"
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

type WordlistInput struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	LanguageCode string `json:"languageCode" binding:"required"`
}

type WordlistsRoutes struct{}
type Wordlist = service.Wordlist

func (h *WordlistsRoutes) GetAll(c *gin.Context) {
	var userId int64 = c.GetInt64("userID")

	wordlists, err := service.GetUserWordlists(userId)
	if err != nil {
		panic(err)
	}
	c.IndentedJSON(http.StatusOK, wordlists)
}

func (h *WordlistsRoutes) Create(c *gin.Context) {
	var input WordlistInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userId int64 = c.GetInt64("userID")
	saved, err := service.SaveWordlist(&Wordlist{Name: input.Name, Description: input.Description, UserID: userId, LanguageCode: input.LanguageCode})

	if err != nil {
		panic(err)
	}

	c.IndentedJSON(http.StatusCreated, saved)
}

func (h *WordlistsRoutes) GetById(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	var userId int64 = c.GetInt64("userID")

	wordlist, err := service.GetWordlistById(id, userId)
	if err != nil {
		if errors.Is(err, &common.NotFoundError{}) {
			c.Status(http.StatusNotFound)
		} else {
			panic(err)
		}
		return
	}
	c.IndentedJSON(http.StatusOK, wordlist)
}

func (h *WordlistsRoutes) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	var userId int64 = c.GetInt64("userID")

	_, err := service.DeleteWordlist(id, userId)
	if err != nil {
		if errors.Is(err, &common.NotFoundError{}) {
			c.Status(http.StatusNotFound)
		} else {
			panic(err)
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *WordlistsRoutes) Update(c *gin.Context) {
	var input WordlistInput

	id, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userId int64 = c.GetInt64("userID")
	err := service.UpdateWordlist(&Wordlist{ID: id, Name: input.Name, Description: input.Description, LanguageCode: input.LanguageCode, UserID: userId})
	if err != nil {
		if errors.Is(err, common.NotFoundError{}) {
			c.Status(http.StatusNotFound)
		} else {
			panic(err)
		}
		return
	}
	c.Status(http.StatusNoContent)
}
