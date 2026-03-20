package postgres

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
)

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w", domainerrors.ErrNotFound)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%w", domainerrors.ErrAlreadyExists)
		case "23503", "23514":
			return fmt.Errorf("%w", domainerrors.ErrConflict)
		}
	}
	return err
}
