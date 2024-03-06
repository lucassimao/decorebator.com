package words

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"github.com/gin-gonic/gin"
)

type WordInput struct {
	Name string `json:"name" binding:"required"`
}

type WordsHandlers struct{}

var Handlers = &WordsHandlers{}

func (h *WordsHandlers) GetAll(c *gin.Context) {
	wordlistId, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	userId := c.GetInt64("userID")

	words, err := GetWordsFromWordlist(wordlistId, userId)
	if err != nil {
		log.Println("Error in GetAll:", err)
		c.String(http.StatusInternalServerError, "Couldn't get user words")
		return
	}
	c.IndentedJSON(http.StatusOK, words)
}

func (h *WordsHandlers) Create(c *gin.Context) {
	wordlistId, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	userId := c.GetInt64("userID")
	var input WordInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	saved, err := SaveWord(&Word{Name: input.Name, UserID: userId, WordlistID: wordlistId})

	if err != nil {
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusCreated, saved)
	}
}

func (h *WordsHandlers) Delete(c *gin.Context) {
	userId := c.GetInt64("userID")
	id, _ := strconv.ParseInt(c.Param("wordId"), 10, 64)

	_, err := DeleteWord(id, userId)
	if err != nil {
		if errors.Is(err, &common.NotFoundError{}) {
			c.String(http.StatusNotFound, err.Error())
		} else {
			c.String(http.StatusInternalServerError, "Couldn't delete wordlist #%d", id)
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *WordsHandlers) Update(c *gin.Context) {
	var input WordInput

	id, _ := strconv.ParseInt(c.Param("wordId"), 10, 64)
	userId := c.GetInt64("userID")

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := UpdateWord(&Word{ID: id, Name: input.Name, UserID: userId})
	if err != nil {
		if errors.Is(err, common.NotFoundError{}) {
			c.String(http.StatusNotFound, err.Error())
		} else {
			c.String(http.StatusInternalServerError, "Couldn't update word #%d", id)
		}
		return
	}
	c.Status(http.StatusNoContent)
}
