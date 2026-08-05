package unit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"decorebator.com/internal/common"
	httpapi "decorebator.com/internal/http"
	"decorebator.com/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type quizAnswerContractStrategy struct {
	result common.QuizResult
	err    error
}

func (*quizAnswerContractStrategy) CreateQuiz(context.Context, int64, int64, []model.QuizType) (*model.Quiz, error) {
	return nil, errors.New("unused")
}

func (s *quizAnswerContractStrategy) SaveQuizResult(_ context.Context, result common.QuizResult, _ bool, _ *pgx.Tx) error {
	s.result = result
	return s.err
}

func quizAnswerRequest(t *testing.T, strategy common.SpacedRepetitionStrategy, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("userID", int64(42))
	context.Request = httptest.NewRequest(http.MethodPatch, "/wordlists/7/quizzes", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	httpapi.NewQuizRoutes(strategy, nil).Save(context)
	context.Writer.WriteHeaderNow()
	return recorder
}

func TestQuizAnswerContractRemainsAnEmptyPerAnswerCommand(t *testing.T) {
	strategy := &quizAnswerContractStrategy{}
	response := quizAnswerRequest(t, strategy, `{
		"wordlistID":7,
		"wordID":11,
		"definitionID":13,
		"leitnerSystemTrackingID":17,
		"quizType":"multiple_choice",
		"isCorrect":true,
		"responseTimeMs":900
	}`)

	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Empty(t, response.Body.String(), "the answer command must not imply a server-owned session summary")
	assert.Equal(t, int64(42), strategy.result.UserID)
	assert.Equal(t, int64(7), strategy.result.WordlistID)
	assert.True(t, strategy.result.IsCorrect)
	assert.Equal(t, 900, strategy.result.ResponseTimeMs)
}

func TestQuizAnswerContractKeepsRequestFailuresOutOfSessionAggregation(t *testing.T) {
	for _, test := range []struct {
		name       string
		strategy   *quizAnswerContractStrategy
		body       string
		wantStatus int
	}{
		{name: "invalid input", strategy: &quizAnswerContractStrategy{}, body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "tracking not found", strategy: &quizAnswerContractStrategy{err: common.NotFoundError{ID: 17, Entity: "leitner_system_tracking"}}, body: `{"wordlistID":7,"wordID":11,"definitionID":13,"leitnerSystemTrackingID":17,"quizType":"multiple_choice","isCorrect":false,"responseTimeMs":900}`, wantStatus: http.StatusNotFound},
		{name: "storage failure", strategy: &quizAnswerContractStrategy{err: errors.New("database unavailable")}, body: `{"wordlistID":7,"wordID":11,"definitionID":13,"leitnerSystemTrackingID":17,"quizType":"multiple_choice","isCorrect":false,"responseTimeMs":900}`, wantStatus: http.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := quizAnswerRequest(t, test.strategy, test.body)
			assert.Equal(t, test.wantStatus, response.Code)
			require.NotEmpty(t, response.Body.String())
		})
	}
}
