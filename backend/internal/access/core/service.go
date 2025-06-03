package core

import (
	"context"

	"vpainless/internal/pkg/authz"
	"vpainless/internal/pkg/db"
)

type usersRepository interface {
	GetUser(ctx context.Context, id UserID, partial authz.Clause) (*User, error)
	SaveUser(ctx context.Context, user *User, partial authz.Clause) (*User, error)
	// should only be used by authorize, others should use get user
	FindUserByName(ctx context.Context, username string) (*User, error)
	ListUsers(ctx context.Context, partial authz.Clause) ([]*User, error)
}

type groupsRepository interface {
	GetGroup(ctx context.Context, id GroupID) (*Group, error)
	SaveGroup(ctx context.Context, group *Group) (*Group, error)
}

type hostingAdapter interface {
	NotifyGroupCreated(ctx context.Context, g *Group) error
}

type AccessRepository interface {
	db.Transactor
	usersRepository
	groupsRepository
}

type Service struct {
	enforcer *authz.Validator
	repo     AccessRepository
	hosting  hostingAdapter
}

func NewService(repo AccessRepository, adapter hostingAdapter) *Service {
	var opts []authz.ValidatorOption
	for path, content := range policies() {
		opts = append(opts, authz.WithRegoModule(path, content))
	}

	return &Service{
		enforcer: authz.NewValidator("access", opts...),
		repo:     repo,
		hosting:  adapter,
	}
}
