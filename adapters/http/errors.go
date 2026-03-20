package httpapi

import (
	"errors"

	"github.com/danielgtaylor/huma/v2"
	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
)

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domainerrors.ErrNotFound):
		return huma.Error404NotFound(err.Error())
	case errors.Is(err, domainerrors.ErrInvalidInput):
		return huma.Error422UnprocessableEntity(err.Error())
	case errors.Is(err, domainerrors.ErrAlreadyExists):
		return huma.Error409Conflict(err.Error())
	case errors.Is(err, domainerrors.ErrUnauthorized):
		return huma.Error401Unauthorized(err.Error())
	case errors.Is(err, domainerrors.ErrForbidden):
		return huma.Error403Forbidden(err.Error())
	case errors.Is(err, domainerrors.ErrConflict):
		return huma.Error409Conflict(err.Error())
	case errors.Is(err, domainerrors.ErrTokenBudgetExceeded):
		return huma.Error429TooManyRequests(err.Error())
	case errors.Is(err, domainerrors.ErrUnsupportedFileType):
		return huma.Error422UnprocessableEntity(err.Error())
	case errors.Is(err, domainerrors.ErrKnowledgeLimitExceeded):
		return huma.Error422UnprocessableEntity(err.Error())
	default:
		var validationErr *domainerrors.ValidationError
		if errors.As(err, &validationErr) {
			return huma.Error422UnprocessableEntity(err.Error())
		}
		return huma.Error500InternalServerError("internal error")
	}
}
