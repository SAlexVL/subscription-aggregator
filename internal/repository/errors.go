package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("subscription not found")

// mapDBError converts driver-level errors into repository domain errors.
func mapDBError(op string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return fmt.Errorf("%s: %w", op, err)
}
