package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	"vpainless/internal/hosting/core"
	"vpainless/internal/pkg/authz"
	"vpainless/internal/pkg/db"
	"vpainless/pkg/querybuilder"
)

func (r *Repository) GetInstance(ctx context.Context, id core.InstanceID, partial authz.Clause) (*core.Instance, error) {
	var result *core.Instance
	if err := r.db.InTxDo(ctx, sql.LevelReadCommitted, func(ctx context.Context, tx *db.Tx) error {
		qb := querybuilder.New(`
			select id, user_id, remote_id, ip, status, connection_str, private_key, created_at
			from instances
		`)

		conds := []querybuilder.Cond{
			{Text: "id = ?", Args: []any{id}},
			{Text: "deleted_at is null"},
		}
		if !partial.IsNil() {
			conds = append(conds, querybuilder.Condition(partial.Condition, partial.Values))
		}
		qb.Where(conds...)
		query, args := qb.SQL()
		row := tx.QueryRowContext(ctx, query, args...)
		var err error
		result, err = scanInstance(row)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return core.ErrNotFound
			}

			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func scanInstance(row Scanner) (*core.Instance, error) {
	var (
		result           core.Instance
		ip               sql.NullString
		createdAt        string
		connectionString sql.NullString
	)

	err := row.Scan(&result.ID, &result.Owner, &result.RemoteID, &ip, &result.Status, &connectionString, &result.PrivateKey, &createdAt)
	if err != nil {
		return nil, err
	}

	if ip.Valid {
		result.IP = net.ParseIP(ip.String)
	}
	if connectionString.Valid {
		result.Config.ConnectionString = connectionString.String
	}
	result.CreatedAt, err = time.Parse(time.DateTime, createdAt)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *Repository) SaveInstance(ctx context.Context, instance *core.Instance) (*core.Instance, error) {
	if err := r.db.InTxDo(ctx, sql.LevelSerializable, func(ctx context.Context, tx *db.Tx) error {
		createdAt := instance.CreatedAt.Format(time.DateTime)
		updatedAt := time.Now().Format(time.DateTime)

		qb := querybuilder.New(`
			insert into instances (
				id,
				user_id,
				remote_id,
				ip,
				status,
				connection_str,
				private_key,
				created_at,
				updated_at
			) values (?, ?, ?, nullif(?, ''), ?, nullif(?, ''), ?, ?, ?)
			on conflict (id) do update set
				ip = excluded.ip,
				status = excluded.status,
				connection_str = excluded.connection_str,
				updated_at = ?
			where deleted_at is null;
		`, instance.ID, instance.Owner, instance.RemoteID, instance.IP.String(), string(instance.Status),
			instance.Config.ConnectionString, instance.PrivateKey, createdAt, updatedAt, updatedAt,
		)
		query, args := qb.SQL()
		_, err := tx.ExecContext(ctx, query, args...)
		return err
	}); err != nil {
		return nil, err
	}

	return instance, nil
}

func (r *Repository) FindInstance(ctx context.Context, id core.UserID) (*core.Instance, error) {
	var result *core.Instance
	if err := r.db.InTxDo(ctx, sql.LevelReadCommitted, func(ctx context.Context, tx *db.Tx) error {
		qb := querybuilder.New(`
			select id, user_id, remote_id, ip, status, connection_str, private_key, created_at
			from instances
			where user_id = ? and deleted_at is null;
		`, id)
		query, args := qb.SQL()
		row := tx.QueryRowContext(ctx, query, args...)
		var err error
		result, err = scanInstance(row)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return core.ErrNotFound
			}

			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) ListInstances(ctx context.Context, partial authz.Clause) ([]*core.Instance, error) {
	var result []*core.Instance
	if err := r.db.InTxDo(ctx, sql.LevelReadCommitted, func(ctx context.Context, tx *db.Tx) error {
		qb := querybuilder.New(`
			select i.id, i.user_id, i.remote_id, i.ip, i.status, i.connection_str, i.private_key, i.created_at
			from instances i
			inner join users u on u.id = i.user_id
		`)

		conds := []querybuilder.Cond{
			querybuilder.Condition("i.deleted_at is null", nil),
		}
		if !partial.IsNil() {
			conds = append(conds, querybuilder.Condition(partial.Condition, partial.Values))
		}
		qb.Where(conds...)
		query, args := qb.SQL()
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}

		result, err = scanInstances(rows)
		return err
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func scanInstances(rows *sql.Rows) ([]*core.Instance, error) {
	var instances []*core.Instance

	for rows.Next() {
		instance, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}

		instances = append(instances, instance)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return instances, nil
}

func (r *Repository) DeleteInstance(ctx context.Context, id core.InstanceID, partial authz.Clause) error {
	return r.db.InTxDo(ctx, sql.LevelSerializable, func(ctx context.Context, tx *db.Tx) error {
		now := time.Now().Format(time.DateTime)
		qb := querybuilder.New(`update instances set deleted_at = ?`, now)

		conds := []querybuilder.Cond{
			querybuilder.Condition("id = ?", []any{id}),
			querybuilder.Condition("deleted_at is null", nil),
		}

		if !partial.IsNil() {
			conds = append(conds, querybuilder.Condition(partial.Condition, partial.Values))
		}

		qb.Where(conds...)
		query, args := qb.SQL()

		result, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}

		count, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if count == 0 {
			return core.ErrNotFound
		}

		if count > 1 {
			return fmt.Errorf("multiple rows deleted")
		}

		return nil
	})
}
