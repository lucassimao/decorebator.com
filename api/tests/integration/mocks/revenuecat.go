package mocks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// RevenueCatMock is a mock HTTP server that simulates the RevenueCat API.
type RevenueCatMock struct {
	Server               *httptest.Server
	CustomerInfoResponse map[string]interface{}
}

// NewRevenueCatMock creates a new RevenueCatMock.
func NewRevenueCatMock() *RevenueCatMock {
	mock := &RevenueCatMock{}
	mock.Server = httptest.NewServer(http.HandlerFunc(mock.handle))
	return mock
}

// Close closes the mock server.
func (m *RevenueCatMock) Close() {
	m.Server.Close()
}

// handle is the HTTP handler for the mock server.
func (m *RevenueCatMock) handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v1/subscribers/test_user_123":
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(m.CustomerInfoResponse)
	default:
		http.NotFound(w, r)
	}
}

// MockGetCustomerInfo sets the mock response for the GetCustomerInfo endpoint.
func (m *RevenueCatMock) MockGetCustomerInfo(response map[string]interface{}) {
	m.CustomerInfoResponse = response
}
