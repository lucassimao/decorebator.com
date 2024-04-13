package common

import "fmt"

type BusinessError struct {
	Message string
}

func (e BusinessError) Error() string {
	return e.Message
}

type NotFoundError struct {
	ID     int64
	Entity string
}

func (e NotFoundError) Is(target error) bool {
	_, ok := target.(*NotFoundError)
	return ok
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s #%d not found", e.Entity, e.ID)
}
