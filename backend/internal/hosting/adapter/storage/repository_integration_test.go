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
	path := filepath.Join(dir, "data", "hosting.db")

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
		delete from startup_scripts;
		delete from ssh_keys;
		delete from xray_templates;
		delete from users;
		delete from groups;

		insert into xray_templates (id, base, group_id)
		values ('00000000-0000-0000-0000-111111111122', 'test_content', '00000000-0000-0000-0000-111111111111');

		insert into ssh_keys (id, group_id, remote_id, name, private_key, public_key)
		values ('00000000-0000-0000-0000-111111111133', '00000000-0000-0000-0000-111111111111', null, 'test key', 'private key', 'public key');

		insert into startup_scripts (id, group_id, remote_id, content)
		values ('00000000-0000-0000-0000-111111111144', '00000000-0000-0000-0000-111111111111', null, "#!/bin/bash");

		insert into groups (id, name, provider_name, provider_url, provider_apikey, default_xray_template, default_ssh_key, default_startup_script)
		values (
			'00000000-0000-0000-0000-111111111111',
			'test_group',
			'vultr',
			'https://api.vultr.com',
			'vultr_api_key',
			'00000000-0000-0000-0000-111111111122',
			'00000000-0000-0000-0000-111111111133',
			'00000000-0000-0000-0000-111111111144'
		), (
			'00000000-0000-0000-0000-222222222222',
			'test_group',
			'vultr',
			'https://api.vultr.com',
			'vultr_api_key',
			'00000000-0000-0000-0000-111111111122',
			'00000000-0000-0000-0000-111111111133',
			'00000000-0000-0000-0000-111111111144'
		)
		;

		insert into users (id, group_id, role) values
			('11000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-111111111111', 'admin'),
			('22000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-111111111111', 'client'),
			('33000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-222222222222', 'client');

		insert into instances (
				id,
				user_id,
				remote_id,
				ip,
				status,
				connection_str,
				private_key,
				created_at,
				updated_at,
				deleted_at
			) values ('00000000-1100-0000-0000-000000000000', '11000000-0000-0000-0000-000000000000',
			'00000000-2200-0000-0000-000000000000', '0.0.0.0', 'ok', 'xray://', 'private_key',
			'2006-01-02 15:04:05', '2006-01-02 15:04:05', '2007-01-02 15:04:05');
	`)
	s.Require().NoError(err, "should insert test materials successfully")
	s.Require().NoError(tx.Commit(), "should commit tx successfully")
}

func (s *RepositoryTestSuite) TearDownTest() {
	s.Require().NoError(s.db.Close(), "should close the connection successfully")
}
