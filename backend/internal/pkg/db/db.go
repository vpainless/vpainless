package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type (
	ctxtype string
	Tx      = sql.Tx
	DB      struct {
		*sql.DB
	}
	Conn struct {
		*sql.Conn
	}
)

type TransactionFunc func(ctx context.Context) error

type Transactor interface {
	Transact(ctx context.Context, level sql.IsolationLevel, fn TransactionFunc) error
}

// newConn creates a new sql connection and returns it.
// It setups the connection for the following cases:
//  1. sqlite does not have gen_uuid_v4 function. But we can
//     implement this on the client side.
func (d *DB) newConn(ctx context.Context) (*sql.Conn, error) {
	conn, err := d.Conn(ctx)
	if err != nil {
		return nil, err
	}

	err = conn.Raw(func(driverConn any) error {
		if sqliteConn, ok := driverConn.(interface{ RegisterFunc(string, any, bool) error }); ok {
			return sqliteConn.RegisterFunc("gen_uuid_v4", func() string {
				return uuid.Must(uuid.NewV4()).String()
			}, true)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (d *DB) InTxDo(ctx context.Context, level sql.IsolationLevel, workload func(ctx context.Context, tx *Tx) error) error {
	if tx := ctx.Value(ctxtype("tx")); tx != nil {
		return workload(ctx, tx.(*Tx))
	}

	conn, err := d.newConn(ctx)
	if err != nil {
		return fmt.Errorf("cannot obtain a new connection: %w", err)
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: level})
	if err != nil {
		return fmt.Errorf("cannot begin a transaction: %w", err)
	}

	ctx = context.WithValue(ctx, ctxtype("tx"), tx)
	if err := workload(ctx, tx); err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	return tx.Commit()
}

func OpenDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_foreign_keys=on", path))
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func ApplyMigrations(db *DB, migrationsPath string) error {
	slog.Info("applying migrations...", "path", migrationsPath)
	driver, err := sqlite.WithInstance(db.DB, &sqlite.Config{NoTxWrap: true})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "ql", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
