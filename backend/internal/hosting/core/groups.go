package core

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/gofrs/uuid/v5"
)

type GroupID struct {
	uuid.UUID
}

type ProviderName string

const (
	Vultr ProviderName = "vultr"
)

type Provider struct {
	Base   url.URL
	Name   ProviderName
	APIKey string
}

type StartUpScriptID struct{ uuid.UUID }

type StartUpScript struct {
	ID       StartUpScriptID
	RemoteID StartUpScriptID
	Content  string
}

type Group struct {
	ID                   GroupID
	Name                 string
	Host                 Provider
	DefaultStartUpScript StartUpScript
	DefaultSSHKey        SSHKeyPair
	DefaultXrayTemplate  XrayTemplateID
	XrayTemplates        map[XrayTemplateID]XrayTemplate
}

// CreateGroup updates the group read model in hosting domain
func (s *Service) SaveGroup(ctx context.Context, group *Group) error {
	logger := slog.With("group", group)
	logger.InfoContext(ctx, "hosting create group...")

	return s.repo.Transact(ctx, sql.LevelSerializable, func(ctx context.Context) error {
		if len(group.XrayTemplates) == 0 {
			id := XrayTemplateID{UUID: uuid.Must(uuid.NewV4())}
			group.DefaultXrayTemplate = id
			group.XrayTemplates = map[XrayTemplateID]XrayTemplate{
				id: {
					ID:   id,
					Base: defaultXrayTemplate,
				},
			}
		}

		sshKeyRemoteID, err := s.vps.CreateSSHKey(ctx, group.Host.APIKey, s.systemKey.PublicKey)
		if err != nil {
			return fmt.Errorf("error creating ssh key on vps provider: %w", err)
		}

		group.DefaultSSHKey = s.systemKey
		group.DefaultSSHKey.ID = SSHKeyID{UUID: uuid.Must(uuid.NewV4())}
		group.DefaultSSHKey.RemoteID = sshKeyRemoteID

		scriptRemoteID, err := s.vps.CreateStartupScript(ctx, group.Host.APIKey, s.defaultStartupScript.Content)
		if err != nil {
			return fmt.Errorf("error creating startup script on the vps provider: %w", err)
		}

		group.DefaultStartUpScript = s.defaultStartupScript
		group.DefaultStartUpScript.ID = StartUpScriptID{UUID: uuid.Must(uuid.NewV4())}
		group.DefaultStartUpScript.RemoteID = scriptRemoteID

		_, err = s.repo.SaveGroup(ctx, group)
		return err
	})
}
