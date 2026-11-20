// Package apperrors defines the domain errors shared by the store, service and handler layers.
// Handlers map each type to an HTTP status code.
package apperrors

import "fmt"

// Validation is returned when the input of a request is invalid.
type Validation struct {
	Reason string
}

func (e Validation) Error() string {
	return e.Reason
}

// NotFound is returned when the requested entity does not exist.
type NotFound struct {
	Entity string
	ID     int
}

func (e NotFound) Error() string {
	return fmt.Sprintf("%s with id %d not found", e.Entity, e.ID)
}

// Conflict is returned when an operation would create a duplicate entity.
type Conflict struct {
	Reason string
}

func (e Conflict) Error() string {
	return e.Reason
}
