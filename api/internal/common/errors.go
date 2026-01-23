package common

import "fmt"

// Intended to be "user safe": the details can be safely sent to the end client
type BusinessError struct {
	Message string
}

func (e BusinessError) Error() string {
	return e.Message
}

type QuizUnavailableReason string

const (
	QuizUnavailableWordlistEmpty       QuizUnavailableReason = "wordlist_empty"
	QuizUnavailableWordlistProcessing  QuizUnavailableReason = "wordlist_processing"
	QuizUnavailableNoUnlearnedWords    QuizUnavailableReason = "no_unlearned_words"
	QuizUnavailableNoMatchingQuizTypes QuizUnavailableReason = "no_matching_quiz_types"
)

type QuizUnavailableError struct {
	Reason  QuizUnavailableReason
	Message string
}

func (e QuizUnavailableError) Error() string {
	return e.Message
}

func (e QuizUnavailableError) Is(target error) bool {
	_, ok := target.(QuizUnavailableError)
	return ok
}

type NotFoundError struct {
	ID     int64
	Entity string
	Err    error
}

func (e NotFoundError) Is(target error) bool {
	_, ok := target.(NotFoundError)
	return ok
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s #%d not found", e.Entity, e.ID)
}

func (e NotFoundError) Unwrap() error {
	return e.Err
}

// Internal only. Service layer must handle, log and translate it
// into a BusinessError{"Internal server error"}.
// Should never be sent to the end client
type DatabaseError struct {
	Msg string
	Err error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("DatabaseError: %s, original error: %v", e.Msg, e.Err)
}
