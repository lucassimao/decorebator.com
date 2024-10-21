package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"decorebator.com/internal/common"
	decorebator "decorebator.com/internal/http"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/workers"
	"github.com/h2non/gock"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"go.uber.org/dig"
)

func TestCanReuseDefinition(t *testing.T) {

	// setup injection container
	var container = dig.New()
	container.Provide(workers.NewContentGenerationService)

	// setup http server
	handlers := decorebator.SetupRoutes(container)
	server := httptest.NewServer(handlers)
	defer server.Close()

	// setup http mock
	defer gock.Off() // Flush pending mocks after test execution

	// mock OpenAI
	gock.New("https://api.openai.com").
		Post("/v1/chat/completions").
		Reply(200).
		JSON(chatCompletionResponse)

	// mock Stability AI
	gock.New("https://api.stability.ai").
		Post("/v2beta/stable-image/generate/core").
		Persist().
		Reply(200).
		// 1x1 pixel transparent image
		JSON(`{"image":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/wcAAgUBAOjcxr0AAAAASUVORK5CYII="}`)

	openAIRequestsCount := 0
	stabilityAIRequestsCount := 0
	gock.Observe(func(r *http.Request, m gock.Mock) {
		if r.Method != "POST" {
			return
		}

		if r.URL.Path == "/v1/chat/completions" {
			openAIRequestsCount++
		}

		if r.URL.Path == "/v2beta/stable-image/generate/core" {
			stabilityAIRequestsCount++
		}
	})
	gock.EnableNetworking()
	gock.NetworkingFilter(func(req *http.Request) bool {
		// Only allow real calls to the api server under tests
		return "http://"+req.URL.Host == server.URL
	})

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
	definitionJobFetcher := rivertest.RequireInserted(ctx, t, riverpgxv5.New(db), workers.DefinitionFetcherArgs{}, nil)
	if definitionJobFetcher == nil {
		t.Error("no definition job inserted")
	}

	err := (&workers.DefinitionFetcherWorker{}).Work(ctx, definitionJobFetcher)
	if err != nil {
		t.Error("failed to run definition worker")
	}

	imageGeneratorJobs := rivertest.RequireManyInserted(ctx, t, riverpgxv5.New(db), []rivertest.ExpectedJob{
		{Args: &workers.ImageGeneratorArgs{}},
		{Args: &workers.ImageGeneratorArgs{}},
		{Args: &workers.ImageGeneratorArgs{}},
	})

	if len(imageGeneratorJobs) != 3 {
		t.Errorf("expected 3 image generator jobs. Only %d found", len(imageGeneratorJobs))
	}

	for _, jobRow := range imageGeneratorJobs {
		var args workers.ImageGeneratorArgs
		err = json.Unmarshal(jobRow.EncodedArgs, &args)
		if err != nil {
			t.Error("failed to unmarshal encoded args")
		}
		err = (&workers.ImageGeneratorWorker{}).Work(ctx, &river.Job[workers.ImageGeneratorArgs]{
			Args: args, JobRow: jobRow,
		})
		if err != nil {
			t.Error("failed to run image generator worker")
		}
	}

	// second user
	{
		var user2Auth = testUtils.SignUp(nil)
		var user2Wordlist = testUtils.NewWordlist(user2Auth, nil)

		if user2Wordlist == nil {
			t.Error("failed to create user 2 wordlist")
		}

		var user2Word = testUtils.NewWord(user2Auth, user2Wordlist.ID, decorebator.WordInput{Name: "networking"})
		if user2Word == nil {
			t.Error("failed to create user 2 word")
		}

		// if more than 1 is found, error will be thrown
		rivertest.RequireManyInserted(ctx, t, riverpgxv5.New(db), []rivertest.ExpectedJob{
			{Args: &workers.DefinitionFetcherArgs{}},
		})
	}

	definitionRepository := repository.DefinitionRepository{Db: db}
	name := "networking"
	def1, err := definitionRepository.Find(repository.FindArgs{Name: &name})
	if err != nil {
		t.Error("failed to find definitions by name")
	}
	if len(def1) != 3 {
		t.Errorf("expected only 3 definitions. Found %d", len(def1))
	}

	if openAIRequestsCount != 1 {
		t.Errorf("Expected only 1 call to /v1/chat/completions. %d found", openAIRequestsCount)
	}

	if stabilityAIRequestsCount != 3 {
		t.Errorf("Expected only 3 calls to /v2beta/stable-image/generate/core. %d found", stabilityAIRequestsCount)
	}

}

// response for word networking
var chatCompletionResponse = `{
	"id": "MOCKED",
	"object": "chat.completion",
	"created": 1728380440,
	"model": "gpt-4o-2024-08-06",
	"choices": [
		{
		"index": 0,
		"message": {
			"role": "assistant",
			"content": "{\n  \"results\": [\n    {\n      \"meaning\": \"The action or process of interacting with others to exchange information and develop professional or social contacts.\",\n      \"part_of_speech\": \"noun\",\n      \"examples\": [\n        \"They emphasized the importance of [networking] at the conference.\",\n        \"Effective [networking] can lead to new job opportunities.\",\n        \"[Networking] events are great for meeting industry professionals.\",\n        \"He spent the evening [networking] with other entrepreneurs.\",\n        \"Online [networking] has become increasingly popular.\"\n      ]\n    },\n    {\n      \"meaning\": \"The linking of computers to allow them to operate interactively.\",\n      \"part_of_speech\": \"noun\",\n      \"examples\": [\n        \"The IT department is responsible for [networking] the office.\",\n        \"They specialize in wireless [networking] technologies.\",\n        \"Proper [networking] setup is crucial for seamless data transfer.\",\n        \"[Networking] hardware is essential for a connected infrastructure.\",\n        \"The company invested in new [networking] equipment.\"\n      ]\n    },\n    {\n      \"meaning\": \"The act of creating and maintaining connections online through social media or other platforms for various purposes.\",\n      \"part_of_speech\": \"verb\",\n      \"inflections\": [\n        {\n          \"inflection\": \"networked\",\n          \"tense\": \"past\",\n          \"examples\": [\n            \"She successfully [networked] with influencers for her business.\",\n            \"They [networked] extensively during the trade fair.\",\n            \"He [networked] with many professionals during his career.\"\n          ]\n        },\n        {\n          \"inflection\": \"networked\",\n          \"tense\": \"past participle\",\n          \"examples\": [\n            \"Having [networked] throughout the conference, she gained several new contacts.\",\n            \"They have [networked] with many companies over the years.\",\n            \"He has [networked] to grow his professional circle.\"\n          ]\n        }\n      ]\n    }\n  ]\n}",
			"refusal": null
		},
		"logprobs": null,
		"finish_reason": "stop"
		}
	],
	"usage": {
		"prompt_tokens": 283,
		"completion_tokens": 415,
		"total_tokens": 698,
		"prompt_tokens_details": {
		"cached_tokens": 0
		},
		"completion_tokens_details": {
		"reasoning_tokens": 0
		}
	},
	"system_fingerprint": "fp_2f406b9113"
	}
	`
