package main

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"vpainless/api"

	"github.com/gofrs/uuid/v5"
)

type MockServer struct {
	count atomic.Int32
}

var (
	userID     = uuid.Must(uuid.NewV4())
	adminID    = uuid.Must(uuid.NewV4())
	groupID    = uuid.Must(uuid.NewV4())
	instanceID = uuid.Must(uuid.NewV4())
)

func (s *MockServer) GetMe(w http.ResponseWriter, r *http.Request) {
	user, _, err := basicAuth(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, err)
		return
	}

	writeJSON(w, http.StatusOK, api.User{
		Id:       toPointer(userID),
		Username: toPointer(user),
		Role:     toPointer(api.Admin),
		GroupId:  toUUIDPointer(groupID),
	})
}

func (s *MockServer) GetUser(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	panic("not implemented")
}

func (s *MockServer) PutUser(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	panic("not implemented")
}

func (s *MockServer) PostUser(w http.ResponseWriter, r *http.Request) {
	var req api.PostUserJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, api.User{
		Id:       toPointer(userID),
		Username: req.Username,
		Role:     toPointer(api.Client),
	})
}

func (s *MockServer) ListUsers(w http.ResponseWriter, r *http.Request) {
	_, _, err := basicAuth(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, err)
		return
	}

	users := []api.User{
		{
			GroupId:  toPointer(groupID),
			Id:       toPointer(adminID),
			Role:     toPointer(api.Client),
			Username: toPointer("admin"),
		},
		{
			GroupId:  toPointer(groupID),
			Id:       toPointer(userID),
			Role:     toPointer(api.Client),
			Username: toPointer("john"),
		},
	}

	res := api.Users{
		Count: toPointer(2),
		Users: &users,
	}

	writeJSON(w, http.StatusOK, res)
}

func (s *MockServer) PostGroup(w http.ResponseWriter, r *http.Request) {
	_, _, err := basicAuth(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, err)
		return
	}

	var req api.PostGroupJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, api.Group{
		Id:   toPointer(groupID),
		Name: req.Name,
	})
}

func (s *MockServer) PostInstance(w http.ResponseWriter, r *http.Request) {
	_, _, err := basicAuth(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, err)
		return
	}

	time.Sleep(5 * time.Second)
	result := api.Instance{
		Id:     toPointer(instanceID),
		Owner:  toPointer(userID),
		Ip:     toPointer("110.134.123.5"),
		Status: toPointer(api.Initializing),
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *MockServer) GetInstance(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	_, _, err := basicAuth(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, err)
		return
	}

	status := api.Off
	connectionStr := ""
	s.count.Add(1)
	switch s.count.Load() {
	case 2:
		status = api.Initializing
	case 3:
		status = api.Ok
		connectionStr = "xray://connection-string"
	case 4:
		s.count.Store(0)
	}
	result := api.Instance{
		ConnectionString: toPointer(connectionStr),
		Id:               toPointer(instanceID),
		Owner:            toPointer(userID),
		Ip:               toPointer("110.134.123.5"),
		Status:           toPointer(status),
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *MockServer) ListInstances(w http.ResponseWriter, r *http.Request) {
	oneInstances(w, r)
	// zeroInstances(w, r)
}

func zeroInstances(w http.ResponseWriter, r *http.Request) {
	_, _, err := basicAuth(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, err)
		return
	}

	var result []api.Instance
	writeJSON(w, http.StatusOK, result)
}

func oneInstances(w http.ResponseWriter, r *http.Request) {
	_, _, err := basicAuth(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, err)
		return
	}

	var result []api.Instance
	result = append(result, api.Instance{
		Id:               toPointer(instanceID),
		Owner:            toPointer(userID),
		Ip:               toPointer("110.134.123.5"),
		Status:           toPointer(api.Initializing),
		ConnectionString: toPointer("xray://connnection"),
	})

	writeJSON(w, http.StatusOK, result)
}

func (s *MockServer) DeleteInstance(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	time.Sleep(5 * time.Second)
	writeJSON(w, http.StatusNoContent, nil)
}
