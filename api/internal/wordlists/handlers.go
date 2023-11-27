package wordlists

import (
	"errors"
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"github.com/gin-gonic/gin"
)

type WordlistInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type WordlistsHandlers struct{}

var Handlers = &WordlistsHandlers{}

func (h *WordlistsHandlers) GetAll(c *gin.Context) {
	// TODO
	var userId int64 = -1

	wordlists, err := GetUserWordlists(userId)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, wordlists)
}

func (h *WordlistsHandlers) Create(c *gin.Context) {
	var input WordlistInput

	if err := c.BindJSON(&input); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// should be updated once we support auth
	var userId int64 = -1

	saved, err := SaveWordlist(&Wordlist{Name: input.Name, Description: input.Description, UserID: userId})
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	} else {
		c.IndentedJSON(http.StatusCreated, saved)
	}
}

func (h *WordlistsHandlers) GetById(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	// should be updated once we support auth
	var userId int64 = -1

	wordlist, err := GetWordlistById(id, userId)
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

func (h *WordlistsHandlers) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var userId int64 = -1

	_, err := DeleteWordlist(id, userId)
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

func (h *WordlistsHandlers) Update(c *gin.Context) {
	var input WordlistInput

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := c.BindJSON(&input); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	err := UpdateWordlist(&Wordlist{ID: id, Name: input.Name, Description: input.Description, UserID: -1})
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
