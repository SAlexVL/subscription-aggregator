package repository

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestMapDBError_NoRows(t *testing.T) {
	err := mapDBError("get subscription", pgx.ErrNoRows)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMapDBError_Other(t *testing.T) {
	orig := errors.New("connection reset")
	err := mapDBError("get subscription", orig)
	if errors.Is(err, ErrNotFound) {
		t.Fatal("unexpected ErrNotFound")
	}
	if !errors.Is(err, orig) {
		t.Fatalf("expected wrapped original error, got %v", err)
	}
}

func TestMapDBError_Nil(t *testing.T) {
	if err := mapDBError("op", nil); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
