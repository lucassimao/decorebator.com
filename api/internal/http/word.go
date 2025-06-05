package http

import (
	"errors"
	"net/http"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

type WordInput struct {
	Name    string `json:"name" binding:"required"`
	Notes   string `json:"notes"`
	Learned bool   `json:"learned"`
}

type WordRoutes struct{}

type Word = service.Word

func (h *WordRoutes) GetAll(c *gin.Context) {
	wordlistId, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	userId := c.GetInt64("userID")

	words, err := service.GetWordByWordlist(wordlistId, userId)
	if err != nil {
		common.Logger.Error("failed to get words", "error", err, "userId", userId, "wordlistId", wordlistId)
		c.String(http.StatusInternalServerError, "Could not get user words")
		return
	}
	c.IndentedJSON(http.StatusOK, words)
}

func (h *WordRoutes) Create(ctx *gin.Context) {
	var wordlistId, _ = strconv.ParseInt(ctx.Param("wordlistId"), 10, 64)
	var userId = ctx.GetInt64("userID")
	var input WordInput

	if err := ctx.BindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var saved, err = service.SaveWord(&Word{Name: input.Name, UserID: userId, WordlistID: wordlistId, Notes: input.Notes}, ctx.Request.Context())
	var logger = common.Logger.With("word", input.Name, "userId", userId, "endpoint", ctx.Request.URL.Path)

	if err != nil {
		switch err.(type) {
		case common.BusinessError:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			logger.Error("failed to create word", "error", err)
			ctx.Status(http.StatusInternalServerError)
		}
	} else {
		ctx.JSON(http.StatusCreated, saved)
	}
}

func (h *WordRoutes) Delete(c *gin.Context) {
	userId := c.GetInt64("userID")
	id, _ := strconv.ParseInt(c.Param("wordId"), 10, 64)

	_, err := service.DeleteWord(id, userId)
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

func (h *WordRoutes) Update(c *gin.Context) {
	var input WordInput

	id, _ := strconv.ParseInt(c.Param("wordId"), 10, 64)
	wordlistId, _ := strconv.ParseInt(c.Param("wordlistId"), 10, 64)
	userId := c.GetInt64("userID")

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := service.UpdateWord(&Word{ID: id, Name: input.Name, UserID: userId, Learned: input.Learned, WordlistID: wordlistId}, nil)
	if err != nil {
		if errors.Is(err, common.NotFoundError{}) {
			c.String(http.StatusNotFound, err.Error())
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update word"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *WordRoutes) GetDefinitions(c *gin.Context) {
	userId := c.GetInt64("userID")
	wordId, err := strconv.ParseInt(c.Param("wordId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
		return
	}

	definitions, err := service.GetDefinitionsByWordId(wordId, userId)
	if err != nil {
		common.Logger.Error("failed to get definitions", "error", err, "userId", userId, "wordId", wordId)
		c.String(http.StatusInternalServerError, "Could not get word definitions")
		return
	}

	c.IndentedJSON(http.StatusOK, definitions)
}
