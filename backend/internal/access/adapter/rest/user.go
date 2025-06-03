package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"vpainless/api"
	"vpainless/internal/access/core"
	"vpainless/internal/pkg/authz"

	"github.com/gofrs/uuid/v5"
)

type userService interface {
	GetUser(ctx context.Context, id core.UserID) (*core.User, error)
	UpdateUser(ctx context.Context, user *core.User) (*core.User, error)
	CreateUser(ctx context.Context, user *core.User) (*core.User, error)
	ListUsers(ctx context.Context) ([]*core.User, error)
}

func (a *Adapter) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	principal, err := authz.GetPrincipal(ctx)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, fmt.Errorf("principal not found"))
		return
	}

	user, err := a.service.GetUser(ctx, core.UserID{UUID: principal.ID})
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, core.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, core.ErrUnauthorized):
			status = http.StatusUnauthorized
		}
		slog.ErrorContext(ctx, "error getting user", "error", err)
		writeJSONError(w, status, err)
		return
	}

	writeJSON(w, http.StatusOK, api.User{
		Id:       toPointer(user.ID.UUID),
		Username: toPointer(user.Username),
		Role:     toPointer(api.UserRole(user.Role)),
		GroupId:  toUUIDPointer(user.GroupID.UUID),
	})
}

func (a *Adapter) GetUser(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	ctx := r.Context()
	user, err := a.service.GetUser(ctx, core.UserID{UUID: id})
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, core.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, core.ErrUnauthorized):
			status = http.StatusUnauthorized
		}
		slog.ErrorContext(ctx, "error getting user", "error", err)
		writeJSONError(w, status, err)
		return
	}

	writeJSON(w, http.StatusOK, api.User{
		Id:       toPointer(user.ID.UUID),
		Username: toPointer(user.Username),
		Role:     toPointer(api.UserRole(user.Role)),
		GroupId:  toUUIDPointer(user.GroupID.UUID),
	})
}

func mapCoreUser(apiUser api.User) (core.User, error) {
	role, err := mapCoreRole(apiUser.Role)
	if err != nil {
		return core.User{}, err
	}

	return core.User{
		ID:       core.UserID{UUID: fromPointer(apiUser.Id)},
		GroupID:  core.GroupID{UUID: fromPointer(apiUser.GroupId)},
		Username: fromPointer(apiUser.Username),
		Password: fromPointer(apiUser.Password),
		Role:     role,
	}, nil
}

func mapCoreRole(role *api.UserRole) (core.Role, error) {
	if role == nil {
		return core.Client, nil
	}

	switch *role {
	case api.Admin:
		return core.Admin, nil
	case api.Client:
		return core.Client, nil
	default:
	}

	return core.Client, fmt.Errorf("invalid role %s", *role)
}

func (a *Adapter) PutUser(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var req api.PutUserJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	req.Id = toPointer(id)

	user, err := mapCoreUser(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	ctx := r.Context()
	u, err := a.service.UpdateUser(ctx, &user)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, core.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, core.ErrBadRequest):
			status = http.StatusBadRequest
		case errors.Is(err, core.ErrUnauthorized):
			status = http.StatusUnauthorized
		}
		slog.ErrorContext(ctx, "error editing user", "error", err)
		writeJSONError(w, status, err)
		return
	}

	writeJSON(w, http.StatusOK, api.User{
		Id:       toPointer(u.ID.UUID),
		GroupId:  toUUIDPointer(u.GroupID.UUID),
		Username: toPointer(u.Username),
		Role:     toPointer(api.UserRole(u.Role)),
	})
}

func (a *Adapter) PostUser(w http.ResponseWriter, r *http.Request) {
	var req api.PostUserJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	user, err := mapCoreUser(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	ctx := r.Context()
	u, err := a.service.CreateUser(ctx, &user)
	status := http.StatusCreated
	if err != nil {
		s := http.StatusInternalServerError
		switch {
		case errors.Is(err, core.ErrNotFound):
			s = http.StatusNotFound
		case errors.Is(err, core.ErrBadRequest):
			s = http.StatusBadRequest
		case errors.Is(err, core.ErrUnauthorized):
			s = http.StatusUnauthorized
		}
		if !errors.Is(err, core.ErrAlreadyExists) {
			slog.ErrorContext(ctx, "error creating user", "error", err)
			writeJSONError(w, s, err)
			return
		}

		status = http.StatusOK
	}

	writeJSON(w, status, api.User{
		Id:       toPointer(u.ID.UUID),
		GroupId:  toUUIDPointer(u.GroupID.UUID),
		Username: toPointer(u.Username),
		Role:     toPointer(api.UserRole(u.Role)),
	})
}

func (a *Adapter) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := a.service.ListUsers(ctx)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			writeJSONError(w, http.StatusUnauthorized, err)
			return
		}

		slog.ErrorContext(ctx, "error listing users", "error", err)
		writeJSONError(w, http.StatusInternalServerError, nil)
		return
	}

	var apiUsers []api.User
	for _, u := range users {
		apiUsers = append(apiUsers, api.User{
			Id:       toPointer(u.ID.UUID),
			GroupId:  toUUIDPointer(u.GroupID.UUID),
			Username: toPointer(u.Username),
			Role:     toPointer(api.UserRole(u.Role)),
		})
	}

	res := api.Users{
		Users: &apiUsers,
		Count: toPointer(len(users)),
	}

	writeJSON(w, http.StatusOK, res)
}
