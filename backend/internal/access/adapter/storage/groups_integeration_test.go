package storage

import (
	"context"

	"vpainless/internal/access/core"

	"github.com/gofrs/uuid/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (s *RepositoryTestSuite) Test_GetGroup() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	repo := NewRepository(s.db)

	group := &core.Group{
		ID:   core.GroupID{UUID: uuid.FromStringOrNil("00000000-0000-0000-0000-111111111111")},
		Name: "test_group",
	}

	_, err := repo.GetGroup(ctx, core.GroupID{})
	s.Require().ErrorIs(err, core.ErrNotFound, "should return and error")

	actual, err := repo.GetGroup(ctx, group.ID)
	s.Require().NoError(err, "should successfully get group")
	s.Require().Equal(group, actual, "groups should be equal")
}

func (s *RepositoryTestSuite) Test_SaveGroup() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	repo := NewRepository(s.db)

	group := &core.Group{
		ID:   core.GroupID{UUID: uuid.FromStringOrNil("00000000-0000-0000-0000-333333333333")},
		Name: "test_group_2",
	}

	actual, err := repo.SaveGroup(ctx, group)
	s.Require().NoError(err, "should successfully create group")
	s.Require().Equal(group, actual, "groups should be equal")

	group.Name = "another_name"
	actual, err = repo.SaveGroup(ctx, group)
	s.Require().NoError(err, "should successfully save group")
	s.Require().Equal(group, actual, "groups should be equal")
}
