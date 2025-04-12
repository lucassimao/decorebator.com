package common

import "fmt"

// Intended to be "user safe": the details can be safely sent to the end client
type BusinessError struct {
	Message string
}

func (e BusinessError) Error() string {
	return e.Message
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
