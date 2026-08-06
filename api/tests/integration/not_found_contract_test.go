package integration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWordAndWordlistRoutesUseOwnedNotFoundContract(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	ctx := context.Background()
	ownerToken, ownerID := createWordUpdateUser(t, server)
	otherToken, otherID := createWordUpdateUser(t, server)
	ownerList := createWordUpdateWordlist(ctx, t, server.DB, ownerID, "Owner list")
	ownerEmptyList := createWordUpdateWordlist(ctx, t, server.DB, ownerID, "Owner empty list")
	otherList := createWordUpdateWordlist(ctx, t, server.DB, otherID, "Other list")
	ownerWord := createWordUpdateWord(ctx, t, server.DB, ownerID, ownerList, "owned", "")

	for _, request := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/wordlists/not-a-number"},
		{http.MethodGet, "/wordlists/0"},
		{http.MethodDelete, "/wordlists/not-a-number"},
		{http.MethodDelete, "/wordlists/0"},
		{http.MethodGet, "/wordlists/not-a-number/words"},
		{http.MethodGet, "/wordlists/0/words"},
		{http.MethodPost, "/wordlists/not-a-number/words"},
		{http.MethodPost, "/wordlists/0/words"},
		{http.MethodGet, "/wordlists/not-a-number/words/definitions?ids=1"},
		{http.MethodGet, "/wordlists/0/words/definitions?ids=1"},
		{http.MethodDelete, "/wordlists/1/words/not-a-number"},
		{http.MethodDelete, "/wordlists/1/words/0"},
	} {
		response := performOwnedRequest(t, server.BaseURL, ownerToken, request.method, request.path, "")
		assert.Equal(t, http.StatusBadRequest, response.StatusCode, request.path)
		response.Body.Close()
	}

	absentListPath := "/wordlists/99999999"
	foreignListPath := "/wordlists/" + int64String(otherList)
	assertSameNotFound(t,
		performOwnedResponse(t, server.BaseURL, ownerToken, http.MethodGet, absentListPath, ""),
		performOwnedResponse(t, server.BaseURL, ownerToken, http.MethodGet, foreignListPath, ""),
	)
	assertSameNotFound(t,
		performOwnedResponse(t, server.BaseURL, ownerToken, http.MethodDelete, absentListPath, ""),
		performOwnedResponse(t, server.BaseURL, ownerToken, http.MethodDelete, foreignListPath, ""),
	)

	response := performOwnedRequest(t, server.BaseURL, ownerToken, http.MethodGet, "/wordlists/"+int64String(ownerEmptyList)+"/words", "")
	assert.Equal(t, http.StatusOK, response.StatusCode)
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `[]`, string(body))
	response.Body.Close()
	response = performOwnedRequest(t, server.BaseURL, ownerToken, http.MethodGet, "/wordlists/"+int64String(otherList)+"/words", "")
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	response.Body.Close()
	response = performOwnedRequest(t, server.BaseURL, ownerToken, http.MethodGet, "/wordlists/"+int64String(ownerEmptyList)+"/words/definitions?ids=99999999", "")
	assert.Equal(t, http.StatusOK, response.StatusCode)
	body, err = io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `[]`, string(body))
	response.Body.Close()
	response = performOwnedRequest(t, server.BaseURL, ownerToken, http.MethodGet, "/wordlists/"+int64String(otherList)+"/words/definitions?ids=99999999", "")
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	response.Body.Close()

	response = performOwnedRequest(t, server.BaseURL, ownerToken, http.MethodGet, "/wordlists/"+int64String(ownerList)+"/words/"+int64String(ownerWord)+"/definitions", "")
	assert.Equal(t, http.StatusOK, response.StatusCode)
	response.Body.Close()
	assertSameNotFound(t,
		performOwnedResponse(t, server.BaseURL, ownerToken, http.MethodGet, "/wordlists/"+int64String(ownerList)+"/words/99999999/definitions", ""),
		performOwnedResponse(t, server.BaseURL, otherToken, http.MethodGet, "/wordlists/"+int64String(ownerList)+"/words/"+int64String(ownerWord)+"/definitions", ""),
	)

	absentWordPath := "/wordlists/" + int64String(ownerList) + "/words/99999999"
	foreignWordPath := "/wordlists/" + int64String(ownerList) + "/words/" + int64String(ownerWord)
	assertSameNotFound(t,
		performOwnedResponse(t, server.BaseURL, ownerToken, http.MethodDelete, absentWordPath, ""),
		performOwnedResponse(t, server.BaseURL, otherToken, http.MethodDelete, foreignWordPath, ""),
	)

	assertSameNotFound(t,
		performOwnedResponse(t, server.BaseURL, ownerToken, http.MethodPost, "/wordlists/99999999/words", `{"name":"absent-create"}`),
		performOwnedResponse(t, server.BaseURL, ownerToken, http.MethodPost, "/wordlists/"+int64String(otherList)+"/words", `{"name":"foreign-create"}`),
	)

	var foreignCreateCount int
	err = server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM words WHERE name IN ('foreign-create', 'absent-create')`).Scan(&foreignCreateCount)
	require.NoError(t, err)
	assert.Zero(t, foreignCreateCount)
}

func performOwnedRequest(t *testing.T, baseURL, token, method, path, body string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(method, baseURL+path, strings.NewReader(body))
	require.NoError(t, err)
	request.Header.Set("Authorization", token)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	return response
}

type ownedResponse struct {
	statusCode  int
	contentType string
	body        []byte
}

func performOwnedResponse(t *testing.T, baseURL, token, method, path, requestBody string) ownedResponse {
	t.Helper()
	response := performOwnedRequest(t, baseURL, token, method, path, requestBody)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	return ownedResponse{
		statusCode:  response.StatusCode,
		contentType: response.Header.Get("Content-Type"),
		body:        body,
	}
}

func assertSameNotFound(t *testing.T, first, second ownedResponse) {
	t.Helper()
	assert.Equal(t, http.StatusNotFound, first.statusCode)
	assert.Equal(t, http.StatusNotFound, second.statusCode)
	assert.Equal(t, "application/json; charset=utf-8", first.contentType)
	assert.Equal(t, first.contentType, second.contentType)
	assert.JSONEq(t, string(first.body), string(second.body))
}

func int64String(value int64) string {
	return fmt.Sprintf("%d", value)
}
