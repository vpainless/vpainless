package rest

import (
	"errors"
	"log/slog"
	"net/http"

	"vpainless/api"
	"vpainless/internal/hosting/core"

	"github.com/gofrs/uuid/v5"
)

func (a *Adapter) GetInstance(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	ctx := r.Context()
	instance, err := a.service.GetInstance(ctx, core.InstanceID{UUID: id})
	if err != nil {
		switch {
		case errors.Is(err, core.ErrNotFound):
			writeJSONError(w, http.StatusNotFound, err)
			return
		case errors.Is(err, core.ErrUnauthorized):
			writeJSONError(w, http.StatusUnauthorized, err)
			return
		}

		slog.ErrorContext(ctx, "error getting instance", "error", err)
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, api.Instance{
		ConnectionString: toPointer(instance.Config.ConnectionString),
		Id:               toPointer(instance.ID.UUID),
		Owner:            toPointer(instance.Owner.UUID),
		Ip:               toPointer(instance.IP.String()),
		Status:           toPointer(api.InstanceStatus(instance.Status)),
	})
}

func (a *Adapter) DeleteInstance(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	ctx := r.Context()

	err := a.service.DeleteInstance(ctx, core.InstanceID{UUID: id})
	if err == nil {
		writeJSON(w, http.StatusNoContent, nil)
		return
	}

	switch {
	case errors.Is(err, core.ErrNotFound):
		writeJSONError(w, http.StatusNotFound, err)
		return
	case errors.Is(err, core.ErrUnauthorized):
		writeJSONError(w, http.StatusUnauthorized, err)
		return
	}

	slog.ErrorContext(ctx, "error deleting instance", "error", err)
	writeJSONError(w, http.StatusInternalServerError, err)
}

func (a *Adapter) PostInstance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	instance, err := a.service.CreateInstance(ctx)
	status := http.StatusCreated
	if err != nil {
		s := http.StatusInternalServerError
		switch {
		case errors.Is(err, core.ErrUnauthorized):
			s = http.StatusUnauthorized
		}
		if !errors.Is(err, core.ErrAlreadyExists) {
			slog.ErrorContext(ctx, "error creating instance", "error", err)
			writeJSONError(w, s, err)
			return
		}
		status = http.StatusOK
	}

	writeJSON(w, status, api.Instance{
		ConnectionString: toPointer(instance.Config.ConnectionString),
		Id:               toPointer(instance.ID.UUID),
		Owner:            toPointer(instance.Owner.UUID),
		Ip:               toPointer(instance.IP.String()),
		Status:           toPointer(api.InstanceStatus(instance.Status)),
	})
}

func (a *Adapter) ListInstances(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instances, err := a.service.ListInstances(ctx)
	if err != nil {
		switch {
		case errors.Is(err, core.ErrUnauthorized):
			writeJSONError(w, http.StatusUnauthorized, err)
			return
		}

		slog.ErrorContext(ctx, "error listing instances", "error", err)
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}

	var result []api.Instance
	for _, instance := range instances {
		result = append(result, api.Instance{
			ConnectionString: toPointer(instance.Config.ConnectionString),
			Id:               toPointer(instance.ID.UUID),
			Owner:            toPointer(instance.Owner.UUID),
			Ip:               toPointer(instance.IP.String()),
			Status:           toPointer(api.InstanceStatus(instance.Status)),
		})
	}

	writeJSON(w, http.StatusOK, result)
}
