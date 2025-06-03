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
)

type groupService interface {
	CreateGroup(ctx context.Context, group *core.Group) (*core.Group, error)
}

func (a *Adapter) PostGroup(w http.ResponseWriter, r *http.Request) {
	var req api.PostGroupJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	group, err := mapGroup(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	ctx := r.Context()
	result, err := a.service.CreateGroup(ctx, &group)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, core.ErrBadRequest):
			status = http.StatusBadRequest
		case errors.Is(err, core.ErrUnauthorized):
			status = http.StatusUnauthorized
		}
		slog.ErrorContext(ctx, "error creating group", "error", err)
		writeJSONError(w, status, err)
		return
	}

	g := api.Group{
		Id:   toPointer(result.ID.UUID),
		Name: toPointer(result.Name),
		Vps: &struct {
			Apikey   *string               `json:"apikey,omitempty"`
			Provider *api.GroupVpsProvider `json:"provider,omitempty"`
		}{Provider: toPointer(api.GroupVpsProvider(result.Host))},
	}
	writeJSON(w, http.StatusCreated, g)
}

func mapGroup(g api.Group) (core.Group, error) {
	name := fromPointer(g.Name)
	if name == "" {
		return core.Group{}, fmt.Errorf("group name missing")
	}

	if g.Vps == nil {
		return core.Group{}, fmt.Errorf("vps details missing")
	}

	host := fromPointer(g.Vps.Provider)
	if host == "" {
		return core.Group{}, fmt.Errorf("vps provider missing")
	}

	apikey := fromPointer(g.Vps.Apikey)
	if apikey == "" {
		return core.Group{}, fmt.Errorf("vps api key missing")
	}

	return core.Group{
		ID:     core.GroupID{},
		Name:   name,
		Host:   string(host),
		APIKey: apikey,
	}, nil
}
