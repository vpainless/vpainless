package storage

import (
	"context"
	"database/sql"
	"testing"

	"vpainless/internal/hosting/core"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/require"
)

func (s *RepositoryTestSuite) Test_Get_Create_FindUser() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	user := &core.User{
		Role:    "admin",
		GroupID: core.GroupID{UUID: uuid.FromStringOrNil("00000000-0000-0000-0000-111111111111")},
	}

	repo := NewRepository(s.db)
	_, err := repo.GetUser(ctx, user.ID)
	s.Require().ErrorIs(err, core.ErrNotFound, "should not find the user")

	actual, err := repo.SaveUser(ctx, user)
	s.Require().NoError(err, "should not return any error")
	s.Require().Equal(user, actual, "users should match")

	actual, err = repo.GetUser(ctx, user.ID)
	s.Require().NoError(err, "should return the saved user")
	s.Require().Equal(user, actual, "return user should match")
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
