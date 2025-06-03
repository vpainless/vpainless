package core

import (
	"context"
	"database/sql"
	"errors"

	"vpainless/internal/pkg/authz"

	"github.com/gofrs/uuid/v5"
)

type UserID struct {
	uuid.UUID
}

type Role string

const (
	Admin  Role = "admin"
	Client Role = "client"
)

type User struct {
	ID      UserID
	GroupID GroupID
	Role    Role
}

// savePrincipal saves the current prinicipal in the read model of users
// in hosting domain.
func (s *Service) savePrincipal(ctx context.Context) error {
	principal, err := authz.GetPrincipal(ctx)
	if err != nil {
		return err
	}

	return s.repo.Transact(ctx, sql.LevelSerializable, func(ctx context.Context) error {
		_, err := s.repo.GetUser(ctx, UserID{principal.ID})
		if err == nil {
			return nil
		}

		if !errors.Is(err, ErrNotFound) {
			return err
		}

		_, err = s.repo.SaveUser(ctx, &User{
			ID:      UserID{principal.ID},
			GroupID: GroupID{principal.GroupID},
			Role:    Role(principal.Role),
		})
		return err
	})
}
