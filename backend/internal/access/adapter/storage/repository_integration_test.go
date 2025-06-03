package storage

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"vpainless/internal/pkg/db"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/stretchr/testify/suite"
)

const (
	timeout time.Duration = 15 * time.Second
)

type RepositoryTestSuite struct {
	suite.Suite
	migrationsPath string
	db             *db.DB
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (s *RepositoryTestSuite) SetupSuite() {
	cwd, err := os.Getwd()
	s.Require().NoError(err, "should get current working directory successfully")
	s.migrationsPath = path.Join(cwd, "../../../pkg/db/migrations")
	s.T().Logf("CWD : %s\n", cwd)
}

func (s *RepositoryTestSuite) SetupTest() {
	dir := s.T().TempDir()
	s.Require().NoError(os.Chdir(dir), "should change directory successfully")
	s.Require().NoError(os.Mkdir("data", 0o755), "should make directory successfully")
	path := filepath.Join(dir, "data", "access.db")

	s.T().Logf("DB path : %s\n", path)
	s.T().Logf("Migrations path : %s\n", s.migrationsPath)

	var err error
	s.db, err = db.OpenDB(path)
	s.Require().NoError(err, "should open the database successfully")
	driver, err := sqlite.WithInstance(s.db.DB, &sqlite.Config{NoTxWrap: true})
	s.Require().NoError(err, "should create an sqlite instance successfully")

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", s.migrationsPath), "ql", driver)
	s.Require().NoError(err, "should create find migrations successfully")
	s.Require().NoError(m.Up(), "should apply migrations uccessfully")

	tx, err := s.db.Begin()
	s.Require().NoError(err, "should begin the transaction successfully")
	_, err = tx.Exec(`
		delete from users;
		delete from groups;

		insert into groups (id, name) values ('00000000-0000-0000-0000-111111111111', 'test_group');

		insert into users (id, group_id, username, password, role) values
		('11111111-0000-0000-0000-000000000000', '00000000-0000-0000-0000-111111111111', 'user_1', 'password', 'admin'),
		('22222222-0000-0000-0000-000000000000', '00000000-0000-0000-0000-111111111111', 'user_2', 'password', 'client'),
		('33333333-0000-0000-0000-000000000000', null, 'user_3', 'password', 'client');
	`)
	s.Require().NoError(err, "should insert test materials successfully")
	s.Require().NoError(tx.Commit(), "should commit tx successfully")
}

func (s *RepositoryTestSuite) TearDownTest() {
	s.Require().NoError(s.db.Close(), "should close the connection successfully")
}
