package storage

import (
	"context"
	"database/sql"
	"time"

	"vpainless/internal/pkg/db"

	"github.com/gofrs/uuid/v5"
)

type Scanner interface {
	Scan(dest ...any) error
}

type Repository struct {
	db.Transactor
	db  *db.DB
	now func() time.Time
}

func NewRepository(db *db.DB) *Repository {
	return &Repository{
		db:  db,
		now: time.Now,
	}
}

func (r *Repository) Transact(ctx context.Context, level sql.IsolationLevel, workload db.TransactionFunc) error {
	return r.db.InTxDo(ctx, level, func(ctx context.Context, _ *db.Tx) error {
		return workload(ctx)
	})
}

func uuidOrNull(anID uuid.UUID) sql.NullString {
	if anID.IsNil() {
		return sql.NullString{}
	}

	return sql.NullString{
		String: anID.String(),
		Valid:  anID != uuid.Nil,
	}
}
