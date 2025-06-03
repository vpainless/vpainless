package api

//go:generate go tool oapi-codegen -config config.yaml api.yaml

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
)

type AccessRestAdapter interface {
	GetMe(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request, id uuid.UUID)
	PutUser(w http.ResponseWriter, r *http.Request, id uuid.UUID)
	ListUsers(w http.ResponseWriter, r *http.Request)
	PostUser(w http.ResponseWriter, r *http.Request)
	PostGroup(w http.ResponseWriter, r *http.Request)
}

type HostingRestAdapter interface {
	GetInstance(w http.ResponseWriter, r *http.Request, id UUID)
	DeleteInstance(w http.ResponseWriter, r *http.Request, id UUID)
	ListInstances(w http.ResponseWriter, r *http.Request)
	PostInstance(w http.ResponseWriter, r *http.Request)
}

type Server struct {
	access  AccessRestAdapter
	hosting HostingRestAdapter
}

func NewServer(access AccessRestAdapter, hosting HostingRestAdapter) *Server {
	return &Server{access: access, hosting: hosting}
}

func (s *Server) GetMe(w http.ResponseWriter, r *http.Request) {
	s.access.GetMe(w, r)
}

func (s *Server) GetUser(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	s.access.GetUser(w, r, id)
}

func (s *Server) PutUser(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	s.access.PutUser(w, r, id)
}

func (s *Server) PostUser(w http.ResponseWriter, r *http.Request) {
	s.access.PostUser(w, r)
}

func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request) {
	s.access.ListUsers(w, r)
}

func (s *Server) PostGroup(w http.ResponseWriter, r *http.Request) {
	s.access.PostGroup(w, r)
}

func (s *Server) PostInstance(w http.ResponseWriter, r *http.Request) {
	s.hosting.PostInstance(w, r)
}

func (s *Server) GetInstance(w http.ResponseWriter, r *http.Request, id UUID) {
	s.hosting.GetInstance(w, r, id)
}

func (s *Server) DeleteInstance(w http.ResponseWriter, r *http.Request, id UUID) {
	s.hosting.DeleteInstance(w, r, id)
}

func (s *Server) ListInstances(w http.ResponseWriter, r *http.Request) {
	s.hosting.ListInstances(w, r)
}
