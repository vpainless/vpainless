package storage

import (
	"context"
	"database/sql"

	"vpainless/internal/pkg/authz"
	"vpainless/internal/pkg/db"
	"vpainless/pkg/querybuilder"

	"github.com/gofrs/uuid/v5"
)

type Repository struct {
	db.Transactor
	db *db.DB
}

func NewRepository(db *db.DB) *Repository {
	return &Repository{
		db: db,
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

func fromPolicyPartial(pc authz.Clause) querybuilder.Cond {
	return querybuilder.Cond{
		Text: pc.Condition,
		Args: pc.Values,
	}
}

type Scanner interface {
	Scan(dest ...any) error
}
