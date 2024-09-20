package tests

import (
	"net/http/httptest"
	"testing"

	decorebator "decorebator.com/internal/http"
)

func TestCanReuseDefinition(t *testing.T) {

	handlers := decorebator.SetupHandlers()
	server := httptest.NewServer(handlers)
	defer server.Close()

	testUtils := TestUtils{t, server.URL}

	var auth = testUtils.SignUp(nil)
	var wordlist = testUtils.NewWordlist(auth, nil)

	if wordlist == nil {
		t.Error("no wordlist")
	}

}
