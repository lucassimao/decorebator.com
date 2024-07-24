package http

import (
	"errors"
	"net/http"
	"strconv"

	"decorebator.com/internal/api"
	"decorebator.com/internal/common"
	"github.com/gin-gonic/gin"
)

type wordlistInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type WordlistsRoutes struct{}
type Wordlist = api.Wordlist

func (h *WordlistsRoutes) GetAll(c *gin.Context) {
	var userId int64 = c.GetInt64("userID")

	wordlists, err := api.GetUserWordlists(userId)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, wordlists)
}

func (h *WordlistsRoutes) Create(c *gin.Context) {
	var input wordlistInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userId int64 = c.GetInt64("userID")

	saved, err := api.SaveWordlist(&Wordlist{Name: input.Name, Description: input.Description, UserID: userId})
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	} else {
		c.IndentedJSON(http.StatusCreated, saved)
	}
}

func (h *WordlistsRoutes) GetById(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	var userId int64 = c.GetInt64("userID")

	wordlist, err := api.GetWordlistById(id, userId)
	if err != nil {
		if errors.Is(err, &common.NotFoundError{}) {
			c.Status(http.StatusNotFound)
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.IndentedJSON(http.StatusOK, wordlist)
}

func (h *WordlistsRoutes) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	var userId int64 = c.GetInt64("userID")

	_, err := api.DeleteWordlist(id, userId)
	if err != nil {
		if errors.Is(err, &common.NotFoundError{}) {
			c.Status(http.StatusNotFound)
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *WordlistsRoutes) Update(c *gin.Context) {
	var input wordlistInput

	id, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userId int64 = c.GetInt64("userID")
	err := api.UpdateWordlist(&Wordlist{ID: id, Name: input.Name, Description: input.Description, UserID: userId})
	if err != nil {
		if errors.Is(err, common.NotFoundError{}) {
			c.Status(http.StatusNotFound)
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.Status(http.StatusNoContent)
}
