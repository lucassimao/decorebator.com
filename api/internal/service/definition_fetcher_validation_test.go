package service

import (
	"reflect"
	"testing"

	"decorebator.com/internal/model"
)

func TestValidateDefinitionsPreservesNonVerbRules(t *testing.T) {
	t.Parallel()

	valid := &model.Definition{
		Token:        "harbor",
		Language:     "en",
		Meaning:      "a sheltered body of water",
		PartOfSpeech: "noun",
		Examples:     []string{"The ship entered the [harbor]."},
	}
	if got := validateDefinitions("harbor", []*model.Definition{valid}); len(got) != 0 {
		t.Fatalf("valid definition returned errors: %v", got)
	}

	invalid := &model.Definition{}
	want := []string{
		"definition 1: missing part_of_speech",
		"definition 1: missing meaning",
		"definition 1: missing language",
		"definition 1: missing token",
		"definition 1: non-verb missing examples",
	}
	if got := validateDefinitions("harbor", []*model.Definition{invalid}); !reflect.DeepEqual(got, want) {
		t.Fatalf("validation errors changed\nwant: %v\n got: %v", want, got)
	}
}

func TestValidateDefinitionsPreservesVerbRuleOrdering(t *testing.T) {
	t.Parallel()

	definition := &model.Definition{
		Token:        "walk",
		Language:     "en",
		Meaning:      "to move on foot",
		PartOfSpeech: "verb",
		Examples:     []string{"[walk]"},
		Inflections: []model.Inflection{
			{Tense: "present", Examples: []string{"without brackets", "[shared]"}},
			{Tense: "past", Inflection: "walked", Examples: []string{"[shared]"}},
		},
	}
	want := []string{
		"definition 1: verb/phrasal verb must have empty examples array but has 1 examples",
		"definition 1: inflection missing form",
		"definition 1: inflection example missing brackets: without brackets",
		"definition 1: duplicate example across inflections '[shared]' found in tenses: [present past]",
		"definition 1: verb missing past participle inflection",
	}
	if got := validateDefinitions("walk", []*model.Definition{definition}); !reflect.DeepEqual(got, want) {
		t.Fatalf("validation errors changed\nwant: %v\n got: %v", want, got)
	}
}
