package http

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestParsePageRequestRejectsInvalidBounds(t *testing.T) {
	t.Parallel()

	for _, target := range []string{"/?limit=0", "/?limit=101", "/?limit=invalid", "/?cursor=0", "/?cursor=-2", "/?cursor=invalid"} {
		t.Run(target, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, target, nil)
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = request

			_, err := parsePageRequest(context)
			require.Error(t, err)
		})
	}
}

func TestPageItemsExposesKeysetCursorOnlyWhenMoreRowsExist(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	page := pageRequest{Limit: 2}

	items := pageItems(context, page, []int64{9, 8, 7}, func(item int64) int64 { return item })
	require.Equal(t, []int64{9, 8}, items)
	require.Equal(t, "8", recorder.Header().Get(nextCursorHeader))

	items = pageItems(context, page, []int64{7}, func(item int64) int64 { return item })
	require.Equal(t, []int64{7}, items)
	require.Empty(t, recorder.Header().Get(nextCursorHeader))
}

func TestParseDefinitionBatchWordIDsRejectsDuplicatesAndExcess(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"1,1", "1,,2", "1,nope", ",1", ""} {
		t.Run(input, func(t *testing.T) {
			_, err := parseDefinitionBatchWordIDs(input)
			require.Error(t, err)
		})
	}
	_, err := parseDefinitionBatchWordIDs(string(make([]byte, maxDefinitionBatchIDsQuerySize+1)))
	require.Error(t, err)

	ids, err := parseDefinitionBatchWordIDs("3,2,1")
	require.NoError(t, err)
	require.Equal(t, []int64{3, 2, 1}, ids)

	overstated := make([]byte, 0, maxDefinitionBatchWordIDs*2+1)
	for i := 0; i <= maxDefinitionBatchWordIDs; i++ {
		if i > 0 {
			overstated = append(overstated, ',')
		}
		overstated = append(overstated, '1')
	}
	_, err = parseDefinitionBatchWordIDs(string(overstated))
	require.Error(t, err)
}

func TestDefinitionBatchContinuationCursorsAreBoundedAndScopedToRequestedIDs(t *testing.T) {
	t.Parallel()
	encoded := encodeDefinitionBatchCursors(map[int64]int64{3: 31, 7: 71, 9: 0})
	cursors, err := parseDefinitionBatchCursors(encoded, []int64{3, 7, 9})
	require.NoError(t, err)
	require.Equal(t, map[int64]int64{3: 31, 7: 71, 9: 0}, cursors)

	for _, raw := range []string{
		"invalid",
		encodeDefinitionBatchCursors(map[int64]int64{9: 91}),
		encodeDefinitionBatchCursors(map[int64]int64{3: -1}),
	} {
		_, parseErr := parseDefinitionBatchCursors(raw, []int64{3, 7})
		require.Error(t, parseErr)
	}
}

func TestCooldownCursorIsOpaqueAndStrictlyValidated(t *testing.T) {
	t.Parallel()
	cursor := service.ErrorReportCooldownCursor{
		CooldownUntil: time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC),
		WordID:        3,
		DefinitionID:  4,
		ErrorType:     "wrong_definition",
	}

	parsed, err := parseCooldownCursor(encodeCooldownCursor(cursor))
	require.NoError(t, err)
	require.Equal(t, &cursor, parsed)

	for _, raw := range []string{"invalid", base64.RawURLEncoding.EncodeToString([]byte(`{"wordId":1}`))} {
		_, parseErr := parseCooldownCursor(raw)
		require.Error(t, parseErr)
	}

	unknownField, err := json.Marshal(map[string]interface{}{
		"cooldownUntil": cursor.CooldownUntil,
		"wordId":        cursor.WordID,
		"definitionId":  cursor.DefinitionID,
		"errorType":     cursor.ErrorType,
		"extra":         true,
	})
	require.NoError(t, err)
	_, err = parseCooldownCursor(base64.RawURLEncoding.EncodeToString(unknownField))
	require.Error(t, err)
}

func TestMasteryAndSubscriptionCursorsRoundTripWithoutChangingSortKeys(t *testing.T) {
	t.Parallel()
	lastSeenAt := time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC)

	mastery, err := parseMasteryCursor(encodeMasteryPageCursor(0.7, &lastSeenAt, 9))
	require.NoError(t, err)
	require.Equal(t, "0.70", mastery.MasteryLevel)
	require.Equal(t, &lastSeenAt, mastery.LastSeenAt)
	require.Equal(t, int64(9), mastery.WordID)

	createdAt := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.UTC)
	subscription, err := parseSubscriptionCursor(encodeSubscriptionCursor(&model.Subscription{ID: 7, CreatedAt: createdAt}))
	require.NoError(t, err)
	require.Equal(t, createdAt, subscription.CreatedAt)
	require.Equal(t, int64(7), subscription.ID)

	for _, raw := range []string{
		"invalid",
		encodeOpaqueCursor(masteryCursorPayload{MasteryLevel: "1.10", WordID: 1}),
		encodeOpaqueCursor(masteryCursorPayload{MasteryLevel: "0.700", WordID: 1}),
	} {
		_, parseErr := parseMasteryCursor(raw)
		require.Error(t, parseErr)
	}
}
