package vpsprovider

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"slices"
	"strings"

	"vpainless/internal/hosting/core"
	"vpainless/pkg/collect"
	"vpainless/pkg/vultr"

	"github.com/gofrs/uuid/v5"
)

type Key = string

const (
	tagVpainless        = "vpainless"
	VpainlessScriptName = "vpainless-script"
	VpainlessKeyName    = "vpainless-publickey"
)

type Vultr struct {
	client         *vultr.Client
	sshKeys        collect.Map[Key, []vultr.SSHKeyID]
	startupScripts collect.Map[Key, vultr.StartupScript]
}

func NewVultr(host url.URL) *Vultr {
	return &Vultr{
		client:         vultr.NewClient(host),
		sshKeys:        collect.Map[Key, []vultr.SSHKeyID]{},
		startupScripts: collect.Map[Key, vultr.StartupScript]{},
	}
}

func (v *Vultr) CreateStartupScript(ctx context.Context, apikey Key, content string) (core.StartUpScriptID, error) {
	ctx = v.client.WithAPIKey(ctx, apikey)

	list, err := v.client.ListStartupScripts(ctx)
	if err != nil {
		return core.StartUpScriptID{}, err
	}

	base64Content := base64.StdEncoding.EncodeToString([]byte(content))
	index := slices.IndexFunc(list.Scripts, func(e vultr.StartupScript) bool {
		if e.Name != VpainlessScriptName {
			return false
		}

		script, err := v.client.GetStartupScript(ctx, e.ID)
		if err != nil {
			slog.WarnContext(ctx, "error getting startup script", "id", e.ID, "error", err)
			return false
		}

		return script.Script == base64Content
	})
	if index >= 0 {
		slog.InfoContext(ctx, "found startup script", "script", list.Scripts[index])
		return core.StartUpScriptID{UUID: uuid.FromStringOrNil(string(list.Scripts[index].ID))}, nil
	}

	req := vultr.CreateStartupScriptRequest{
		Name:   "vpainless-script",
		Type:   vultr.Boot,
		Script: base64Content,
	}

	script, err := v.client.CreateStartupScript(ctx, &req)
	if err != nil {
		return core.StartUpScriptID{UUID: uuid.Nil}, err
	}

	slog.InfoContext(ctx, "created startup script", "script", script)
	return core.StartUpScriptID{UUID: uuid.FromStringOrNil(string(script.ID))}, nil
}

func (v *Vultr) CreateSSHKey(ctx context.Context, apikey Key, publickey []byte) (core.SSHKeyID, error) {
	ctx = v.client.WithAPIKey(ctx, apikey)

	list, err := v.client.ListSSHKeys(ctx)
	if err != nil {
		return core.SSHKeyID{}, err
	}
	index := slices.IndexFunc(list.SSHKeys, func(e vultr.SSHKey) bool {
		if e.Name != VpainlessKeyName {
			return false
		}

		key, err := v.client.GetSSHKey(ctx, e.ID)
		if err != nil {
			slog.WarnContext(ctx, "error getting ssh key", "id", e.ID, "error", err)
			return false
		}

		return strings.TrimSpace(key.PublicKey) == strings.TrimSpace(string(publickey))
	})
	if index >= 0 {
		slog.InfoContext(ctx, "found ssh key", "key", list.SSHKeys[index])
		return core.SSHKeyID{UUID: uuid.FromStringOrNil(string(list.SSHKeys[index].ID))}, nil
	}

	req := vultr.CreateSSHKeyRequest{
		Name: VpainlessKeyName,
		Key:  string(publickey),
	}

	key, err := v.client.CreateSSHKey(ctx, req)
	if err != nil {
		return core.SSHKeyID{UUID: uuid.Nil}, err
	}

	return core.SSHKeyID{UUID: uuid.FromStringOrNil(string(key.ID))}, nil
}

func (v *Vultr) CreateInstance(ctx context.Context, apikey Key, param core.CreateInstanceParam) (*core.RemoteInstance, error) {
	ctx = v.client.WithAPIKey(ctx, apikey)
	keyID, err := v.loadSSHKeyID(ctx, apikey, param)
	if err != nil {
		return nil, err
	}

	scriptID, err := v.loadScriptID(ctx, apikey, param)
	if err != nil {
		return nil, err
	}

	req := vultr.CreateInstanceRequest{
		Region:   vultr.Frankfurt,
		Plan:     vultr.BasicPlan,
		OS:       vultr.Debian12,
		Label:    param.Label,
		Backup:   vultr.BackupDisabled,
		SSHKeys:  []vultr.SSHKeyID{keyID},
		Tags:     []string{tagVpainless},
		ScriptID: toPointer(string(scriptID)),
	}

	instance, err := v.client.CreateInstance(ctx, req)
	if err != nil {
		return nil, err
	}

	return &core.RemoteInstance{
		ID: core.InstanceID{UUID: instance.ID},
		IP: net.ParseIP(instance.MainIP),
	}, nil
}

// GetInstance returns an information about an instance from vultr.
// The only reliable information are instance's ip and creation date.
// Everything else should be retrived from somewhere else.
func (v *Vultr) GetInstance(ctx context.Context, apikey Key, id core.InstanceID) (*core.RemoteInstance, error) {
	ctx = v.client.WithAPIKey(ctx, apikey)
	instance, err := v.client.GetInstance(ctx, vultr.InstanceID(id.UUID))
	if err != nil {
		return nil, err
	}

	return &core.RemoteInstance{
		ID: core.InstanceID{UUID: instance.ID},
		IP: net.ParseIP(instance.MainIP),
	}, nil
}

// DeleteInstance deletes a instance from vultr.
func (v *Vultr) DeleteInstance(ctx context.Context, apikey Key, id core.InstanceID) error {
	ctx = v.client.WithAPIKey(ctx, apikey)
	return v.client.DeleteInstance(ctx, vultr.InstanceID(id.UUID))
}

// loadSSHKeyID adds the ssh key if not already present to vultr account, otherwise, it returns it's id
func (v *Vultr) loadSSHKeyID(ctx context.Context, apikey Key, param core.CreateInstanceParam) (vultr.SSHKeyID, error) {
	keyIDs, ok := v.sshKeys.Load(apikey)
	if !ok {
		slog.InfoContext(ctx, "initializing vultr ssh-keys")
		resp, err := v.client.ListSSHKeys(ctx)
		if err != nil {
			return "", fmt.Errorf("error listing ssh keys: %w", err)
		}

		keyIDs = make([]vultr.SSHKeyID, 0, len(resp.SSHKeys))
		for _, key := range resp.SSHKeys {
			keyIDs = append(keyIDs, key.ID)
		}

		v.sshKeys.Store(apikey, keyIDs)
	}

	id := vultr.SSHKeyID(param.SSHKey.RemoteID.String())
	if slices.Contains(keyIDs, id) {
		return id, nil
	}

	slog.InfoContext(ctx, "creating public key...")
	key, err := v.client.CreateSSHKey(ctx, vultr.CreateSSHKeyRequest{
		Name: param.SSHKey.Name,
		Key:  string(param.SSHKey.PublicKey), // TODO: this might cause issue
	})
	if err != nil {
		return "", fmt.Errorf("error creating vultr ssh key: %w", err)
	}

	keyIDs = append(keyIDs, key.ID)
	v.sshKeys.Store(apikey, keyIDs)

	return key.ID, nil
}

func (v *Vultr) loadScriptID(ctx context.Context, apikey Key, param core.CreateInstanceParam) (vultr.StartupScriptID, error) {
	script, ok := v.startupScripts.Load(apikey)
	if !ok {
		slog.InfoContext(ctx, "initializing vultr startup scripts")
		resp, err := v.client.ListStartupScripts(ctx)
		if err != nil {
			return "", fmt.Errorf("error listing ssh keys: %w", err)
		}

		index := slices.IndexFunc(resp.Scripts, func(e vultr.StartupScript) bool {
			return vultr.StartupScriptID(param.Script.RemoteID.String()) == e.ID
		})
		if index >= 0 {
			script = resp.Scripts[index]
			v.startupScripts.Store(apikey, script)
			return script.ID, nil
		}
	}

	if script.ID != "" {
		return script.ID, nil
	}

	slog.InfoContext(ctx, "creating startup script...")
	ss, err := v.client.CreateStartupScript(ctx, &vultr.CreateStartupScriptRequest{
		Name:   "vpainless-script",
		Type:   vultr.Boot,
		Script: param.Script.Content,
	})
	if err != nil {
		return "", fmt.Errorf("error creating vultr startup script: %w", err)
	}

	v.startupScripts.Store(apikey, *ss)
	return script.ID, nil
}

func toPointer[T any](v T) *T {
	return &v
}
