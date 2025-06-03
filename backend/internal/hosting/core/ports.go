package core

import (
	"context"

	"vpainless/internal/pkg/authz"
	"vpainless/internal/pkg/db"
)

type VPSProvider interface {
	CreateInstance(ctx context.Context, apikey string, param CreateInstanceParam) (*RemoteInstance, error)
	GetInstance(ctx context.Context, apikey string, id InstanceID) (*RemoteInstance, error)
	DeleteInstance(ctx context.Context, apikey string, id InstanceID) error
	CreateSSHKey(ctx context.Context, apikey string, publickey []byte) (SSHKeyID, error)
	CreateStartupScript(ctx context.Context, apikey string, content string) (StartUpScriptID, error)
}

type Repository interface {
	db.Transactor
	userRepository
	groupRepository
	instanceRepository
}

type userRepository interface {
	GetUser(ctx context.Context, id UserID) (*User, error)
	SaveUser(ctx context.Context, user *User) (*User, error)
}

type groupRepository interface {
	GetGroup(ctx context.Context, id GroupID) (*Group, error)
	SaveGroup(ctx context.Context, group *Group) (*Group, error)
}

type instanceRepository interface {
	GetInstance(ctx context.Context, id InstanceID, partial authz.Clause) (*Instance, error)
	DeleteInstance(ctx context.Context, id InstanceID, partial authz.Clause) error
	FindInstance(ctx context.Context, id UserID) (*Instance, error)
	ListInstances(ctx context.Context, partial authz.Clause) ([]*Instance, error)
	SaveInstance(ctx context.Context, instance *Instance) (*Instance, error)
}
