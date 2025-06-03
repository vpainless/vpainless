package core

import (
	"context"
	"database/sql"
	"errors"

	"vpainless/internal/pkg/authz"

	"github.com/gofrs/uuid/v5"
)

const (
	ResourceGroups  = "groups"
	VultrApiKeySize = 36
)

type GroupID struct{ uuid.UUID }

type Group struct {
	ID     GroupID
	Name   string
	Host   string
	APIKey string
}

func (s *Service) CreateGroup(ctx context.Context, g *Group) (*Group, error) {
	principal, err := authz.GetPrincipal(ctx)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if len(g.APIKey) != VultrApiKeySize {
		return nil, errors.Join(ErrBadRequest, errors.New("invalid api key"))
	}

	pls, err := s.enforcer.Can(ctx, principal, authz.Create, authz.Resource{Group: ResourceGroups})
	if err != nil || !pls.Allow {
		return nil, ErrUnauthorized
	}

	g.ID = GroupID{uuid.Must(uuid.NewV4())}
	result := g
	// TODO: This is a bad idea, we cannot put this inside the transaction
	// because it calles the service directly and while using access db transction.
	// but this causes a conflict, because hosting also has a group table.
	//
	// We cannot put this outside the transaction either,
	// because everything goes fine on hosting side, but we fail to commit
	// the transaction, we will end up with a stale group in the hosting core.
	//
	// For now I just moved it outside the transaction to fail fast.
	if err := s.hosting.NotifyGroupCreated(ctx, result); err != nil {
		return nil, err
	}

	if err := s.repo.Transact(ctx, sql.LevelSerializable, func(ctx context.Context) error {
		result, err = s.repo.SaveGroup(ctx, g)
		if err != nil {
			return err
		}

		user, err := s.repo.GetUser(ctx, UserID{principal.ID}, authz.Clause{})
		if err != nil {
			return err
		}

		user.GroupID = g.ID
		user.Role = Admin
		_, err = s.repo.SaveUser(ctx, user, authz.Clause{})
		return err
	}); err != nil {
		return nil, err
	}

	return result, nil
}
