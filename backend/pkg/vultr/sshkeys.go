package vultr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SSHKeyID string

type SSHKeysResponse struct {
	SSHKeys []SSHKey `json:"ssh_keys"`
}

type SSHKey struct {
	ID        SSHKeyID  `json:"id"`
	CreateAt  time.Time `json:"date_created"`
	Name      string    `json:"name"`
	PublicKey string    `json:"ssh_key"`
}

func (c *Client) ListSSHKeys(ctx context.Context) (*SSHKeysResponse, error) {
	res, err := c.do(ctx, http.MethodGet, "v2/ssh-keys", nil)
	if err != nil {
		return nil, fmt.Errorf("error listings ssh-keys: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("error listings ssh-keys: %s %s", http.StatusText(res.StatusCode), body)
	}

	resp := SSHKeysResponse{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

type CreateSSHKeyRequest struct {
	Name string `json:"name"`
	Key  string `json:"ssh_key"`
}

type SSHKeyResponse struct {
	SSHKey `json:"ssh_key"`
}

func (c *Client) CreateSSHKey(ctx context.Context, req CreateSSHKeyRequest) (*SSHKey, error) {
	res, err := c.do(ctx, http.MethodPost, "v2/ssh-keys", req)
	if err != nil {
		return nil, fmt.Errorf("error creating ssh-key: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("error creating ssh-key: %s %s", http.StatusText(res.StatusCode), body)
	}

	resp := SSHKeyResponse{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp.SSHKey, nil
}

func (c *Client) GetSSHKey(ctx context.Context, id SSHKeyID) (*SSHKey, error) {
	res, err := c.do(ctx, http.MethodGet, fmt.Sprintf("v2/ssh-keys/%s", id), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting ssh key: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("error getting ssh key: status: %d %s msg: %s", res.StatusCode, http.StatusText(res.StatusCode), string(body))
	}

	resp := SSHKeyResponse{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("error decoding request: %w", err)
	}

	return &resp.SSHKey, nil
}
