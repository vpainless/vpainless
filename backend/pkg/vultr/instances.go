package vultr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofrs/uuid/v5"
)

type (
	InstanceID   = uuid.UUID
	PlanID       string
	RegionID     string
	OSID         int
	BackupStatus string
	PowerStatus  string
	Status       string
	ServerStatus string
)

const (
	BasicPlan      PlanID       = "vc2-1c-1gb"
	Debian12       OSID         = 2136
	Frankfurt      RegionID     = "fra"
	Warsaw         RegionID     = "waw"
	BackupEnabled  BackupStatus = "enabled"
	BackupDisabled BackupStatus = "disabled"

	StatusActive    Status = "active"
	StatusPending   Status = "pending"
	StatusSuspended Status = "suspended"
	StatusResizing  Status = "resizing"

	PowerStatusRunning PowerStatus = "running"
	PowerStatusStopped PowerStatus = "stopped"

	ServerStatusNone              ServerStatus = "none"
	ServerStatusLocked            ServerStatus = "locked"
	ServerStatusInstallingBooting ServerStatus = "installingbooting"
	ServerStatusOK                ServerStatus = "ok"
)

type Instance struct {
	ID           InstanceID   `json:"id"`
	OS           string       `json:"os"`
	MainIP       string       `json:"main_ip"`
	DateCreated  time.Time    `json:"date_created"`
	Label        string       `json:"label"`
	Tags         []string     `json:"tags"`
	Status       Status       `json:"status"`
	PowerStatus  PowerStatus  `json:"power_status"`
	ServerStatus ServerStatus `json:"server_status"`
}

type ListInstancesResponse struct {
	Instances []Instance `json:"instances"`
}

func (c *Client) ListInstances(ctx context.Context) (*ListInstancesResponse, error) {
	res, err := c.do(ctx, http.MethodGet, "v2/instances", nil)
	if err != nil {
		return nil, fmt.Errorf("error listings instances: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("error listing instances: %s %s", http.StatusText(res.StatusCode), body)
	}

	resp := ListInstancesResponse{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

type CreateInstanceRequest struct {
	Region   RegionID     `json:"region"`
	Plan     PlanID       `json:"plan"`
	OS       OSID         `json:"os_id"`
	Label    string       `json:"label"`
	Tags     []string     `json:"tags,omitempty"`
	SSHKeys  []SSHKeyID   `json:"sshkey_id"`
	ScriptID *string      `json:"script_id,omitempty"`
	Backup   BackupStatus `json:"backups"`
}

type CreateInstanceResponse struct {
	Instance `json:"instance"`
}

func (c *Client) CreateInstance(ctx context.Context, req CreateInstanceRequest) (*CreateInstanceResponse, error) {
	res, err := c.do(ctx, http.MethodPost, "v2/instances", &req)
	if err != nil {
		return nil, fmt.Errorf("error creating instance: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("error creating instance: status: %d %s msg: %s", res.StatusCode, http.StatusText(res.StatusCode), string(body))
	}

	resp := &CreateInstanceResponse{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("error decoding request: %w", err)
	}

	return resp, err
}

type GetInstanceResponse struct {
	Instance `json:"instance"`
}

func (c *Client) GetInstance(ctx context.Context, id InstanceID) (*Instance, error) {
	res, err := c.do(ctx, http.MethodGet, fmt.Sprintf("v2/instances/%s", id), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting instance: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("error getting instance: status: %d %s msg: %s", res.StatusCode, http.StatusText(res.StatusCode), string(body))
	}

	resp := GetInstanceResponse{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("error decoding request: %w", err)
	}

	return &resp.Instance, nil
}

func (c *Client) DeleteInstance(ctx context.Context, id InstanceID) error {
	res, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("v2/instances/%s", id), nil)
	if err != nil {
		return fmt.Errorf("error deleting instance ipv4: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNoContent {
		return nil
	}

	if res.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	body, _ := io.ReadAll(res.Body)
	return fmt.Errorf("error deleting instance: status: %d %s msg: %s", res.StatusCode, http.StatusText(res.StatusCode), string(body))
}
