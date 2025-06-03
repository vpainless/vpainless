package core

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"vpainless/internal/pkg/authz"
	"vpainless/pkg/middleware"

	"github.com/gofrs/uuid/v5"
)

type UserID struct{ uuid.UUID }

type Role string

const (
	Client        Role = "client"
	Admin         Role = "admin"
	ResourceUsers      = "users"
)

type User struct {
	ID       UserID  `json:"id"`
	GroupID  GroupID `json:"group_id"`
	Username string  `json:"username"`
	Password string  `json:"password"`
	Role     Role    `json:"role"`
}

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrBadRequest    = errors.New("bad request")
)

func generatePassword(username, password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(username+password)))
}

func (s *Service) GetUser(ctx context.Context, id UserID) (*User, error) {
	slog.Info("getting user", "id", id)
	principal, err := authz.GetPrincipal(ctx)
	if err != nil {
		return nil, errors.Join(err, ErrUnauthorized)
	}

	policy, err := s.enforcer.Can(ctx, principal, authz.Get, authz.ResourceID(ResourceUsers, id.UUID))
	if err != nil || !policy.Allow {
		return nil, errors.Join(err, ErrUnauthorized)
	}

	user, err := s.repo.GetUser(ctx, id, policy.Partial)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrNotFound
	}

	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, newUser *User) (*User, error) {
	var result *User
	principal, err := authz.GetPrincipal(ctx)
	if err != nil {
		return nil, ErrUnauthorized
	}

	err = s.repo.Transact(ctx, sql.LevelSerializable, func(ctx context.Context) error {
		oldUser, err := s.repo.GetUser(ctx, newUser.ID, authz.Clause{})
		if err != nil {
			return err
		}

		if newUser.Username == "" {
			newUser.Username = oldUser.Username
		}
		if newUser.Password == "" {
			newUser.Password = oldUser.Password
		} else {
			newUser.Password = generatePassword(newUser.Username, newUser.Password)
		}

		input := map[string]any{
			"id":       newUser.ID,
			"new_role": newUser.Role,
			"old_role": oldUser.Role,
		}
		if !newUser.GroupID.IsNil() {
			input["new_group_id"] = newUser.GroupID
		}
		if !oldUser.GroupID.IsNil() {
			input["old_group_id"] = oldUser.GroupID
		}

		policy, err := s.enforcer.Can(ctx, principal, authz.Update, authz.Resource{
			Group: ResourceUsers,
			Value: input,
		})

		if err != nil || !policy.Allow {
			return errors.Join(err, ErrUnauthorized)
		}

		result, err = s.repo.SaveUser(ctx, newUser, policy.Partial)
		return err
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) ListUsers(ctx context.Context) ([]*User, error) {
	principal, err := authz.GetPrincipal(ctx)
	if err != nil {
		return nil, ErrUnauthorized
	}

	var result []*User
	err = s.repo.Transact(ctx, sql.LevelSerializable, func(ctx context.Context) error {
		policy, err := s.enforcer.Can(ctx, principal, authz.List, authz.Resource{
			Group: ResourceUsers,
			Value: nil,
		})

		if err != nil || !policy.Allow {
			return errors.Join(err, ErrUnauthorized)
		}

		result, err = s.repo.ListUsers(ctx, policy.Partial)
		return err
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) CreateUser(ctx context.Context, u *User) (*User, error) {
	var result *User
	var exists bool

	if err := s.repo.Transact(ctx, sql.LevelSerializable, func(ctx context.Context) error {
		var err error
		result, err = s.repo.FindUserByName(ctx, u.Username)
		if err == nil {
			exists = true
			return nil
		}

		user := &User{
			ID:       UserID{uuid.Must(uuid.NewV4())},
			GroupID:  u.GroupID,
			Username: u.Username,
			Password: generatePassword(u.Username, u.Password),
			// Role should be always client when creating a new user.
			// They can always promote users later
			Role: Client,
		}

		slog.InfoContext(ctx, "creating user...", "user", user)
		p, _ := authz.GetPrincipal(ctx)
		policy, err := s.enforcer.Can(ctx, p, authz.Create, authz.ResourceFunc(func() (string, any) {
			if !user.GroupID.IsNil() {
				return ResourceUsers, map[string]any{"group_id": user.GroupID}
			}
			return ResourceUsers, nil
		}))
		if err != nil || !policy.Allow {
			return errors.Join(err, ErrUnauthorized)
		}

		result, err = s.repo.SaveUser(ctx, user, policy.Partial)
		return err
	}); err != nil {
		return nil, err
	}

	if exists {
		return result, ErrAlreadyExists
	}
	return result, nil
}

func (s *Service) Authenticate(ctx context.Context, creds middleware.Credentials) (authz.Principal, error) {
	user, err := s.repo.FindUserByName(ctx, creds.Username)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			err = ErrUnauthorized
		}
		return authz.Principal{}, err
	}

	passwordHash := generatePassword(creds.Username, creds.Password)
	if passwordHash != user.Password {
		// TODO: remove this
		slog.DebugContext(ctx, "authorize failed", "hash", passwordHash)
		return authz.Principal{}, ErrUnauthorized
	}

	return authz.Principal{
		ID:      user.ID.UUID,
		GroupID: user.GroupID.UUID,
		Role:    authz.Role(user.Role),
	}, nil
}
