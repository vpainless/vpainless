package storage

import (
	"context"
	"database/sql"
	"errors"

	"vpainless/internal/hosting/core"
	"vpainless/internal/pkg/db"
	"vpainless/pkg/querybuilder"
)

func (r *Repository) GetUser(ctx context.Context, id core.UserID) (*core.User, error) {
	var result core.User

	err := r.db.InTxDo(ctx, sql.LevelReadCommitted, func(ctx context.Context, tx *db.Tx) error {
		qb := querybuilder.New(`
			select id, group_id, role from users where id = ?	
		`, id)
		query, args := qb.SQL()
		row := tx.QueryRowContext(ctx, query, args...)
		if err := row.Scan(&result.ID, &result.GroupID, &result.Role); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return core.ErrNotFound
			}

			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// SaveUser creates a users and return it.
func (r *Repository) SaveUser(ctx context.Context, user *core.User) (*core.User, error) {
	var result core.User

	err := r.db.InTxDo(ctx, sql.LevelSerializable, func(ctx context.Context, tx *db.Tx) error {
		qb := querybuilder.New(`
				insert into users (id, group_id, role)
				values (?, ?, ?)
				on conflict (id) do update set
					group_id = excluded.group_id,
					role = excluded.role
 				returning id, group_id, role;
		`, user.ID, uuidOrNull(user.GroupID.UUID), user.Role)
		query, args := qb.SQL()
		row := tx.QueryRowContext(ctx, query, args...)
		return row.Scan(&result.ID, &result.GroupID, &result.Role)
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}
