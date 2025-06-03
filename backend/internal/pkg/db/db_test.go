package db

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	dt "github.com/golang-migrate/migrate/v4/database/testing"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	dir := t.TempDir()
	t.Logf("DB path : %s\n", filepath.Join(dir, "sqlite.db"))
	p := &sqlite.Sqlite{}
	addr := fmt.Sprintf("sqlite://%s", filepath.Join(dir, "sqlite.db"))
	d, err := p.Open(addr)
	if err != nil {
		t.Fatal(err)
	}
	dt.Test(t, d, []byte("CREATE TABLE t (Qty int, Name string);"))
}

func TestMigrate(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err, "should get current working directory successfully")
	migrationsPath := path.Join(cwd, "./migrations")

	dir := t.TempDir()
	require.NoError(t, os.Chdir(dir), "should change directory successfully")
	require.NoError(t, os.Mkdir("data", 0o755), "should make directory successfully")
	path := filepath.Join(dir, "data", "access.db")

	t.Logf("DB path : %s\n", path)

	db, err := sql.Open("sqlite", path)
	require.NoError(t, err, "should open the database successfully")
	t.Cleanup(func() {
		require.NoError(t, db.Close(), "should close the connection successfully")
	})
	driver, err := sqlite.WithInstance(db, &sqlite.Config{NoTxWrap: true})
	require.NoError(t, err, "should create an sqlite instance successfully")

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", migrationsPath), "ql", driver)
	require.NoError(t, err, "should create find migrations successfully")
	require.NoError(t, m.Up(), "should up migrations successfully")
	require.NoError(t, m.Down(), "should down migrations successfully")
}
