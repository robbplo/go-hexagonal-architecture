package domainerrors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound               = errors.New("not found")
	ErrAlreadyExists          = errors.New("already exists")
	ErrInvalidInput           = errors.New("invalid input")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrForbidden              = errors.New("forbidden")
	ErrConflict               = errors.New("conflict")
	ErrTokenBudgetExceeded    = errors.New("token budget exceeded")
	ErrUnsupportedFileType    = errors.New("unsupported file type")
	ErrKnowledgeLimitExceeded = errors.New("knowledge token limit exceeded")
	ErrImmutableField         = errors.New("immutable field")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation: %s: %s", e.Field, e.Message)
}
