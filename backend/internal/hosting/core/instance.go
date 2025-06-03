package core

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"vpainless/internal/pkg/authz"
	"vpainless/pkg/remote"
	"vpainless/pkg/vultr"

	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/ssh"
)

type (
	InstanceStatus string
	InstanceID     struct{ uuid.UUID }
	SSHKeyID       struct{ uuid.UUID }
)

const (
	StatusUnknown      InstanceStatus = "unknown"
	StatusOff          InstanceStatus = "off"
	StatusOK           InstanceStatus = "ok"
	StatusInitializing InstanceStatus = "initializing"
)

type RemoteInstance struct {
	ID InstanceID
	IP net.IP
}

type Instance struct {
	ID         InstanceID
	RemoteID   InstanceID
	Owner      UserID
	IP         net.IP
	Status     InstanceStatus
	Config     XrayConfig
	PrivateKey []byte
	CreatedAt  time.Time
}

type SSHKeyPair struct {
	ID         SSHKeyID
	RemoteID   SSHKeyID
	Name       string
	PublicKey  []byte
	PrivateKey []byte
}

type CreateInstanceParam struct {
	SSHKey SSHKeyPair
	Label  string
	Script StartUpScript
}

func (s *Service) GetInstance(ctx context.Context, id InstanceID) (*Instance, error) {
	principal, err := authz.GetPrincipal(ctx)
	if err != nil {
		return nil, ErrUnauthorized
	}

	var instance *Instance
	if err := s.repo.Transact(ctx, sql.LevelReadCommitted, func(ctx context.Context) error {
		policy, err := s.enforcer.Can(ctx, principal, authz.Get, authz.ResourceID(ResourceInstances, id.UUID))
		if err != nil || !policy.Allow {
			return ErrUnauthorized
		}

		instance, err = s.repo.GetInstance(ctx, id, policy.Partial)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return instance, nil
}

func (s *Service) DeleteInstance(ctx context.Context, id InstanceID) error {
	principal, err := authz.GetPrincipal(ctx)
	if err != nil {
		return ErrUnauthorized
	}

	deletePolicy, err := s.enforcer.Can(ctx, principal, authz.Delete, authz.ResourceID(ResourceInstances, id.UUID))
	if err != nil || !deletePolicy.Allow {
		return ErrUnauthorized
	}

	return s.repo.Transact(ctx, sql.LevelSerializable, func(ctx context.Context) error {
		instance, err := s.repo.GetInstance(ctx, id, authz.Clause{})
		if err != nil {
			return err
		}

		if err := s.repo.DeleteInstance(ctx, id, deletePolicy.Partial); err != nil {
			return err
		}

		group, err := s.repo.GetGroup(ctx, GroupID{UUID: principal.GroupID})
		if err != nil {
			return errors.Join(ErrGroups, err)
		}

		err = s.vps.DeleteInstance(ctx, group.Host.APIKey, instance.RemoteID)
		if err == nil {
			return nil
		}

		// If the remote instance is not found, maybe it is deleted manually
		// by any of the admins. The best we can do in this case is to log the
		// error and continue.
		if errors.Is(err, vultr.ErrNotFound) {
			slog.WarnContext(ctx, "remote instance not found", "remote_id", instance.RemoteID)
			return nil
		}
		return err
	})
}

func (s *Service) CreateInstance(ctx context.Context) (*Instance, error) {
	principal, err := authz.GetPrincipal(ctx)
	if err != nil {
		return nil, ErrUnauthorized
	}

	var (
		result *Instance
		apikey string
		param  CreateInstanceParam
	)

	err = s.repo.Transact(ctx, sql.LevelSerializable, func(ctx context.Context) (err error) {
		userID := UserID{UUID: principal.ID}
		policy, err := s.enforcer.Can(ctx, principal, authz.Create, authz.ResourceFunc(func() (string, any) {
			return ResourceInstances, map[string]any{
				"user_id": userID,
			}
		}))
		if err != nil || !policy.Allow {
			return ErrUnauthorized
		}

		if err := s.savePrincipal(ctx); err != nil {
			return err
		}

		result, err = s.repo.FindInstance(ctx, userID)
		if err == nil {
			return nil
		}
		if !errors.Is(err, ErrNotFound) {
			return err
		}

		group, err := s.repo.GetGroup(ctx, GroupID{principal.GroupID})
		if err != nil {
			return err
		}
		apikey = group.Host.APIKey

		param = CreateInstanceParam{
			SSHKey: group.DefaultSSHKey,
			Label:  userID.String()[:8],
			Script: group.DefaultStartUpScript,
		}

		slog.InfoContext(ctx, "creating vultr instance...")
		remoteInstance, err := s.vps.CreateInstance(ctx, apikey, param)
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				err = errors.Join(err, s.vps.DeleteInstance(ctx, apikey, remoteInstance.ID))
			}
		}()

		result = &Instance{
			ID:         InstanceID{UUID: uuid.Must(uuid.NewV4())},
			RemoteID:   remoteInstance.ID,
			Owner:      userID,
			IP:         remoteInstance.IP,
			CreatedAt:  time.Now(),
			Status:     StatusInitializing,
			Config:     XrayConfig{},
			PrivateKey: group.DefaultSSHKey.PrivateKey,
		}

		slog.InfoContext(ctx, "creating db instance...", "user_id", result.Owner, "remote_id", remoteInstance.ID)
		result, err = s.repo.SaveInstance(ctx, result)
		if err != nil {
			return err
		}

		return err
	})
	if err != nil {
		return nil, err
	}

	// TODO: instead of this, we have to read the db and init instances in a separete
	// go routine.
	go s.SetupInstance(apikey, result, param)
	return result, nil
}

func (s *Service) SetupInstance(apikey string, instance *Instance, param CreateInstanceParam) {
	ctx := context.Background()
	err := s.repo.Transact(ctx, sql.LevelSerializable, func(ctx context.Context) error {
		slog.InfoContext(ctx, "core: waiting for instance to finish initialization...", "instance_id", instance.ID.String())
		conn, err := s.waitForSSHClient(ctx, instance, apikey, param.SSHKey.PrivateKey, "root")
		if err != nil {
			return fmt.Errorf("error waiting for ssh client: %w", err)
		}
		defer conn.Close()

		slog.InfoContext(ctx, "core: uploading reality config...", "instance_id", instance.ID)
		realityConfig, err := NewRealityConfig(fakeURL)
		if err != nil {
			return fmt.Errorf("error creating reality config: %w", err)
		}
		b := bytes.NewBufferString(realityConfig.String())
		if err := remote.UploadFile(conn, "/usr/local/etc/xray/config.json", b); err != nil {
			return fmt.Errorf("error uploading reality config: %w", err)
		}

		instance.Config = XrayConfig{
			ConnectionString: realityConfig.ConnectionString(instance.IP),
		}

		slog.WarnContext(ctx, "core: restarting xray...", "instance_id", instance.ID)
		if _, err := remote.Execute(conn, "systemctl restart xray"); err != nil {
			return fmt.Errorf("error restarting xray: %w", err)
		}

		instance.Status = StatusOK

		_, err = s.repo.SaveInstance(ctx, instance)
		if err != nil {
			return fmt.Errorf("error saving instance %s after initialization: %w", instance.ID, err)
		}
		return nil
	})
	if err != nil {
		slog.Error("error setting up instance", "error", err)
	}
}

// waitForSSHClient tries go get a ssh client to the specified instance. It will retries until success
// if instance is still booting up and being initialized.
func (s *Service) waitForSSHClient(ctx context.Context, instance *Instance, apikey string, privateKey []byte, username string) (*ssh.Client, error) {
	var ip net.IP
	var conn *ssh.Client

	log := slog.With("instance", instance.ID)
	ticker := time.NewTicker(time.Second * waitDurationInSeconds)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if ip == nil || ip.IsUnspecified() {
				i, err := s.vps.GetInstance(ctx, apikey, instance.RemoteID)
				if err != nil {
					return nil, err
				}
				ip = i.IP
			}

			if ip == nil || ip.IsUnspecified() {
				log.WarnContext(ctx, "core: no ip assigned to the instance yet...")
				continue
			}

			if conn == nil {
				log.InfoContext(ctx, "core: trying ssh...", "ip", ip)
				var err error
				conn, err = remote.Dial(ip, privateKey, username)
				if err != nil {
					log.WarnContext(ctx, "core: ssh dial error...", "error", err)
					continue
				}
			}

			if _, err := remote.Execute(conn, "which xray"); err != nil {
				log.WarnContext(ctx, "core: ssh command execution error...", "error", err)
				continue
			}

			instance.IP = ip
			return conn, nil
		}
	}
}

func (s *Service) ListInstances(ctx context.Context) ([]*Instance, error) {
	principal, err := authz.GetPrincipal(ctx)
	if err != nil {
		return nil, ErrUnauthorized
	}

	var instances []*Instance
	if err := s.repo.Transact(ctx, sql.LevelReadCommitted, func(ctx context.Context) error {
		policy, err := s.enforcer.Can(ctx, principal, authz.List, authz.Resource{Group: ResourceInstances})
		if err != nil || !policy.Allow {
			return ErrUnauthorized
		}

		instances, err = s.repo.ListInstances(ctx, policy.Partial)
		return err
	}); err != nil {
		return nil, err
	}

	return instances, nil
}
