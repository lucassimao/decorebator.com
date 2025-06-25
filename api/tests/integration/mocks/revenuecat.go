package mocks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// RevenueCatMock is a mock HTTP server that simulates the RevenueCat API.
type RevenueCatMock struct {
	Server               *httptest.Server
	CustomerInfoResponse map[string]interface{}
	mu                   sync.Mutex
}

// NewRevenueCatMock creates a new RevenueCatMock.
func NewRevenueCatMock() *RevenueCatMock {
	mock := &RevenueCatMock{
		CustomerInfoResponse: make(map[string]interface{}),
	}
	mock.Server = httptest.NewServer(http.HandlerFunc(mock.handle))
	return mock
}

// Close closes the mock server.
func (m *RevenueCatMock) Close() {
	m.Server.Close()
}

// GetURL returns the URL of the mock server.
func (m *RevenueCatMock) GetURL() string {
	return m.Server.URL
}

// handle is the HTTP handler for the mock server.
func (m *RevenueCatMock) handle(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if strings.HasPrefix(r.URL.Path, "/v1/subscribers/") {
		w.Header().Set("Content-Type", "application/json")
		// Extract the app_user_id from the path
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 3 {
			appUserID := parts[3]
			// Create a dynamic response
			dynamicResponse := make(map[string]interface{})
			for k, v := range m.CustomerInfoResponse {
				dynamicResponse[k] = v
			}
			dynamicResponse["subscriber"].(map[string]interface{})["original_app_user_id"] = appUserID
			_ = json.NewEncoder(w).Encode(dynamicResponse)
			return
		}
	}
	http.NotFound(w, r)
}

// MockGetCustomerInfo sets the mock response for the GetCustomerInfo endpoint.
func (m *RevenueCatMock) MockGetCustomerInfo(response map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CustomerInfoResponse = response
}
