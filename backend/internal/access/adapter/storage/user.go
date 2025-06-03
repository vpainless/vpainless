package storage

import (
	"context"
	"database/sql"
	"errors"

	"vpainless/internal/access/core"
	"vpainless/internal/pkg/authz"
	"vpainless/internal/pkg/db"
	"vpainless/pkg/querybuilder"

	"github.com/gofrs/uuid/v5"
)

func (r *Repository) GetUser(ctx context.Context, id core.UserID, partial authz.Clause) (*core.User, error) {
	var user *core.User
	exists := true

	err := r.db.InTxDo(ctx, sql.LevelReadCommitted, func(ctx context.Context, tx *db.Tx) error {
		qb := querybuilder.New(
			"select u.id, u.group_id, u.username, u.password, u.role from users u",
		)
		qb.Where(
			querybuilder.Condition("id = ?", []any{id}),
			fromPolicyPartial(partial),
		)

		query, args := qb.SQL()
		row := tx.QueryRowContext(ctx, query, args...)
		var err error
		user, err = scanUser(row)
		if errors.Is(err, sql.ErrNoRows) {
			exists = false
			return nil
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	return user, nil
}

func scanUser(row Scanner) (*core.User, error) {
	var (
		user    core.User
		groupID sql.NullString
	)

	err := row.Scan(&user.ID, &groupID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		return nil, err
	}

	user.GroupID = core.GroupID{UUID: uuid.FromStringOrNil(groupID.String)}

	return &user, nil
}

// SaveUser creates a users and return it. It failes if a user with similar username exists
func (r *Repository) SaveUser(ctx context.Context, user *core.User, partial authz.Clause) (*core.User, error) {
	var result *core.User

	err := r.db.InTxDo(ctx, sql.LevelReadCommitted, func(ctx context.Context, tx *db.Tx) error {
		qb := querybuilder.New(`
				insert into users (id, group_id, username, password, role)
				values (?, ?, ?, ?, ?)
				on conflict (id) do update set
					username = excluded.username,
					password = excluded.password,
					group_id = excluded.group_id,
					role = excluded.role
		`, user.ID, uuidOrNull(user.GroupID.UUID), user.Username, user.Password, user.Role)
		qb.Where(fromPolicyPartial(partial))
		qb.Append(" returning id, group_id, username, password, role;")
		query, args := qb.SQL()
		row := tx.QueryRowContext(ctx, query, args...)
		var err error
		result, err = scanUser(row)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) FindUserByName(ctx context.Context, username string) (*core.User, error) {
	var user *core.User
	err := r.db.InTxDo(ctx, sql.LevelReadCommitted, func(ctx context.Context, tx *db.Tx) error {
		row := tx.QueryRowContext(ctx, "select id, group_id, username, password, role from users where username = ?;", username)
		var err error
		user, err = scanUser(row)
		if errors.Is(err, sql.ErrNoRows) {
			return core.ErrNotFound
		}

		return err
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) ListUsers(ctx context.Context, partial authz.Clause) ([]*core.User, error) {
	var users []*core.User
	err := r.db.InTxDo(ctx, sql.LevelReadCommitted, func(ctx context.Context, tx *db.Tx) error {
		qb := querybuilder.New(
			"select u.id, u.group_id, u.username, u.password, u.role from users u",
		)
		qb.Where(
			fromPolicyPartial(partial),
		)

		query, args := qb.SQL()
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}

		users, err = scanUsers(rows)
		return err
	})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func scanUsers(rows *sql.Rows) ([]*core.User, error) {
	var users []*core.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
