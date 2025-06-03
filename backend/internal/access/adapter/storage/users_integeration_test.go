package storage

import (
	"context"
	"database/sql"
	"testing"

	"vpainless/internal/access/core"
	"vpainless/internal/pkg/authz"

	"github.com/gofrs/uuid/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
)

func (s *RepositoryTestSuite) Test_Get_Create_FindUser() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	user := &core.User{
		Username: "john",
		Password: "hash",
		Role:     "admin",
		GroupID:  core.GroupID{UUID: uuid.FromStringOrNil("00000000-0000-0000-0000-111111111111")},
	}

	repo := NewRepository(s.db)
	nilPartil := authz.Clause{}
	groupPartial := authz.Clause{
		Condition: "group_id = ?",
		Values:    []any{user.GroupID},
	}

	result, err := repo.GetUser(ctx, user.ID, nilPartil)
	s.Require().NoError(err, "should not return any error")
	s.Require().Nil(result, "result should be nil")

	actual, err := repo.SaveUser(ctx, user, nilPartil)
	s.Require().NoError(err, "should not return any error")
	s.Require().Equal(user, actual, "users should match")

	actual, err = repo.GetUser(ctx, user.ID, nilPartil)
	s.Require().NoError(err, "should work with nil partial")
	s.Require().Equal(user, actual, "users should match")

	actual, err = repo.GetUser(ctx, user.ID, groupPartial)
	s.Require().NoError(err, "should work with group partial")
	s.Require().Equal(user, actual, "users should match")

	actual, err = repo.FindUserByName(ctx, user.Username)
	s.Require().NoError(err, "should not return any error")
	s.Require().Equal(user, actual, "users should match")
}

func (s *RepositoryTestSuite) Test_SaveUser_Partials() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	user := &core.User{
		Username: "john",
		Password: "hash",
		Role:     "admin",
		GroupID:  core.GroupID{UUID: uuid.FromStringOrNil("00000000-0000-0000-0000-111111111111")},
	}

	repo := NewRepository(s.db)
	nilPartial := authz.Clause{}
	actual, err := repo.SaveUser(ctx, user, nilPartial)
	s.Require().NoError(err, "should not return any error")
	s.Require().Equal(user, actual, "users should match")

	partial := authz.Clause{
		Condition: "group_id = ?",
		Values:    []any{user.GroupID},
	}
	actual, err = repo.SaveUser(ctx, user, partial)
	s.Require().NoError(err, "should not return any error")
	s.Require().Equal(user, actual, "users should match")

	anotherGroupPartial := authz.Clause{
		Condition: "group_id = ?",
		Values:    []any{uuid.Nil},
	}
	_, err = repo.SaveUser(ctx, user, anotherGroupPartial)
	s.Require().Error(err, "should not return any error")
}

func TestUUIDorNull(t *testing.T) {
	t.Parallel()

	id := uuid.Must(uuid.NewV4())

	tt := []struct {
		name   string
		input  uuid.UUID
		expect sql.NullString
	}{
		{
			name:   "nil uuid",
			input:  uuid.Nil,
			expect: sql.NullString{},
		},
		{
			name:   "valid uuid",
			input:  id,
			expect: sql.NullString{String: id.String(), Valid: true},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := uuidOrNull(tc.input)
			require.Equal(t, tc.expect, actual, "should match")
		})
	}
}

func (s *RepositoryTestSuite) Test_ListUsers() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	repo := NewRepository(s.db)
	users := []*core.User{
		{
			ID:       core.UserID{UUID: uuid.FromStringOrNil("11111111-0000-0000-0000-000000000000")},
			GroupID:  core.GroupID{UUID: uuid.FromStringOrNil("00000000-0000-0000-0000-111111111111")},
			Username: "user_1",
			Password: "password",
			Role:     "admin",
		},
		{
			ID:       core.UserID{UUID: uuid.FromStringOrNil("22222222-0000-0000-0000-000000000000")},
			GroupID:  core.GroupID{UUID: uuid.FromStringOrNil("00000000-0000-0000-0000-111111111111")},
			Username: "user_2",
			Password: "password",
			Role:     "client",
		},
		{
			ID:       core.UserID{UUID: uuid.FromStringOrNil("33333333-0000-0000-0000-000000000000")},
			Username: "user_3",
			Password: "password",
			Role:     "client",
		},
	}

	nilPartil := authz.Clause{}
	actual, err := repo.ListUsers(ctx, nilPartil)
	s.Require().NoError(err, "should not return any error")
	s.Require().ElementsMatch(users, actual, "should list all users")

	adminPartial := authz.Clause{
		Condition: "group_id = ?",
		Values:    []any{uuid.FromStringOrNil("00000000-0000-0000-0000-111111111111")},
	}
	actual, err = repo.ListUsers(ctx, adminPartial)
	s.Require().NoError(err, "should not return any error")
	s.Require().ElementsMatch(users[:2], actual, "should list all users")

	clientPartial := authz.Clause{
		Condition: "id = ?",
		Values:    []any{uuid.FromStringOrNil("22222222-0000-0000-0000-000000000000")},
	}
	actual, err = repo.ListUsers(ctx, clientPartial)
	s.Require().NoError(err, "should not return any error")
	s.Require().ElementsMatch(users[:1], actual, "should list all users")
}
