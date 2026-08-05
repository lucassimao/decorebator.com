package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"decorebator.com/internal/common"
	decorebatorhttp "decorebator.com/internal/http"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capturingQuizStrategy struct {
	allowed []model.QuizType
}

func (s *capturingQuizStrategy) CreateQuiz(_ context.Context, _, _ int64, allowed []model.QuizType) (*model.Quiz, error) {
	s.allowed = append([]model.QuizType(nil), allowed...)
	return &model.Quiz{ID: 1, Type: model.GuessMeaning}, nil
}

func (*capturingQuizStrategy) SaveQuizResult(context.Context, common.QuizResult, bool, *pgx.Tx) error {
	return nil
}

func TestEffectiveSubscriptionMiddlewareEnablesPremiumQuizBehavior(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeMobileIAPStore{access: true}
	access, err := service.NewEffectiveAccessService(store, model.StoreEnvironmentProduction, func() time.Time { return testIAPNow })
	require.NoError(t, err)
	strategy := &capturingQuizStrategy{}
	quiz := decorebatorhttp.NewQuizRoutes(strategy, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", int64(7))
		c.Set("user", &model.User{ID: 7, SubscriptionPlan: model.PlanFree})
		c.Next()
	})
	router.Use(decorebatorhttp.ResolveEffectiveSubscription(access))
	router.GET("/wordlists/:wordlistId/quizzes", quiz.Create)
	request := httptest.NewRequest(http.MethodGet, "/wordlists/11/quizzes?quizTypes=WORD_FROM_AUDIO", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, []model.QuizType{model.WordFromAudio}, strategy.allowed)
}
