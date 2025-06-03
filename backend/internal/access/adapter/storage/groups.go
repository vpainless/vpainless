package storage

import (
	"context"
	"database/sql"
	"errors"

	"vpainless/internal/access/core"
	"vpainless/internal/pkg/db"
	"vpainless/pkg/querybuilder"
)

func (r *Repository) GetGroup(ctx context.Context, id core.GroupID) (*core.Group, error) {
	var group core.Group
	if err := r.db.InTxDo(ctx, sql.LevelReadCommitted, func(ctx context.Context, tx *db.Tx) error {
		qb := querybuilder.New("select id, name from groups where id = ?;", id)
		query, args := qb.SQL()
		row := tx.QueryRowContext(ctx, query, args...)
		if err := row.Scan(&group.ID, &group.Name); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return core.ErrNotFound
			}
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &group, nil
}

func (r *Repository) SaveGroup(ctx context.Context, group *core.Group) (*core.Group, error) {
	if err := r.db.InTxDo(ctx, sql.LevelLinearizable, func(ctx context.Context, tx *db.Tx) error {
		qb := querybuilder.New(`
			insert into groups (id, name) values (?, ?)
			on conflict (id) do update
				set name = excluded.name;
		`, group.ID, group.Name)
		query, args := qb.SQL()
		_, err := tx.ExecContext(ctx, query, args...)
		return err
	}); err != nil {
		return nil, err
	}

	return group, nil
}
