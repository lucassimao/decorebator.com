package tests

import (
	"net/http"
	"testing"

	decorebator "decorebator.com/internal/http"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
)

type TestUtils struct {
	t         *testing.T
	serverURL string
}

// Sign up new user and return authorization token
func (testUtils *TestUtils) SignUp(signup *decorebator.SignupInput) string {

	serverURL := testUtils.serverURL
	t := testUtils.t

	if signup == nil {
		signup = &decorebator.SignupInput{
			FirstName: gofakeit.FirstName(),
			LastName:  gofakeit.LastName(),
			Email:     gofakeit.Email(),
			Password:  gofakeit.Password(true, true, true, true, false, 10),
		}
	}

	e := httpexpect.Default(t, serverURL)

	resp := e.POST("/users").
		WithJSON(signup).
		Expect()

	resp.Status(http.StatusCreated).
		NoContent().
		Cookie("Authorization")

	authorization := resp.Header("Authorization").NotEmpty()

	if authorization == nil {
		t.Fatal("no authorization header found after sign up")
	}

	return authorization.Raw()
}

func (testUtils TestUtils) NewWordlist(authorization string, newWordlist *decorebator.WordlistInput) *decorebator.Wordlist {
	var serverURL = testUtils.serverURL
	var t = testUtils.t

	if newWordlist == nil {
		newWordlist = &decorebator.WordlistInput{}
		gofakeit.Struct(newWordlist)
	}

	e := httpexpect.Default(t, serverURL)

	var wordlist decorebator.Wordlist
	e.POST("/wordlists").
		WithHeader("authorization", authorization).
		WithJSON(newWordlist).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Decode(&wordlist)

	return &wordlist
}
