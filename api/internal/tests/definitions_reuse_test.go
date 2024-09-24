package tests

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"decorebator.com/internal/api"
	"decorebator.com/internal/common"
	decorebator "decorebator.com/internal/http"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"go.uber.org/dig"
	"go.uber.org/mock/gomock"
)

func TestCanReuseDefinition(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// var mockWorkerTrigger = mocks.NewMockWorkerTrigger(ctrl)

	var container = dig.New()
	container.Provide(api.NewContentGenerationService)

	// mockWorkerTrigger.
	// 	EXPECT().
	// 	TriggerDefinitionFetcher(gomock.Any()).
	// 	Times(2)

	// mockWorkerTrigger.
	// 	EXPECT().
	// 	TriggerTextToSpeech(gomock.Any()).
	// 	Times(2)

	handlers := decorebator.SetupHandlers(container)
	server := httptest.NewServer(handlers)
	defer server.Close()

	testUtils := TestUtils{t, server.URL}

	// first user
	{
		var user1Auth = testUtils.SignUp(nil)
		var user1Wordlist = testUtils.NewWordlist(user1Auth, nil)

		if user1Wordlist == nil {
			t.Error("failed to create user 1 wordlist")
		}

		var user1Word = testUtils.NewWord(user1Auth, user1Wordlist.ID, decorebator.WordInput{Name: "networking"})
		if user1Word == nil {
			t.Error("failed to create user 1 word")
		}
	}

	ctx := common.SetDigContainerInContext(context.Background(), container)
	var db, _ = common.GetDBConnection()
	job := rivertest.RequireInserted(ctx, t, riverpgxv5.New(db), &api.DefinitionFetcherArgs{}, nil)
	fmt.Println(job)

	// second user
	// {
	// 	var user2Auth = testUtils.SignUp(nil)
	// 	var user2Wordlist = testUtils.NewWordlist(user2Auth, nil)

	// 	if user2Wordlist == nil {
	// 		t.Error("failed to create user 2 wordlist")
	// 	}

	// 	var user2Word = testUtils.NewWord(user2Auth, user2Wordlist.ID, decorebator.WordInput{Name: "networking"})
	// 	if user2Word == nil {
	// 		t.Error("failed to create user 2 word")
	// 	}
	// }

}
