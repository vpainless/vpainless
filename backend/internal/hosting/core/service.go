package core

import (
	"errors"

	"vpainless/internal/pkg/authz"
)

var (
	ErrGroups        = errors.New("groups error")
	ErrBadRequest    = errors.New("bad request")
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrUnauthorized  = errors.New("unauthorized")
)

const (
	waitDurationInSeconds = 20
	fakeURL               = "www.speedtest.net"
	ResourceInstances     = "instances"
)

type Service struct {
	vps                  VPSProvider
	repo                 Repository
	systemKey            SSHKeyPair
	defaultStartupScript StartUpScript
	enforcer             *authz.Validator
}

func NewService(repo Repository, vps VPSProvider, systemKey SSHKeyPair, startscript StartUpScript) *Service {
	var opts []authz.ValidatorOption
	for path, content := range policies() {
		opts = append(opts, authz.WithRegoModule(path, content))
	}

	return &Service{
		enforcer:             authz.NewValidator("hosting", opts...),
		vps:                  vps,
		repo:                 repo,
		systemKey:            systemKey,
		defaultStartupScript: startscript,
	}
}
