package vultr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type (
	StartupScriptType string
	StartupScriptID   string
)

const (
	Boot StartupScriptType = "boot"
	PXE  StartupScriptType = "pxe"
)

type CreateStartupScriptRequest struct {
	Name   string            `json:"name"`
	Type   StartupScriptType `json:"type"`
	Script string            `json:"script"`
}

type StartupScript struct {
	ID         StartupScriptID   `json:"id"`
	CreatedAt  time.Time         `json:"date_created"`
	ModifiedAt time.Time         `json:"date_modified"`
	Name       string            `json:"name"`
	Type       StartupScriptType `json:"type"`
	Script     string            `json:"script"`
}

type ScriptResponse struct {
	Script StartupScript `json:"startup_script"`
}

type ScriptsResponse struct {
	Scripts []StartupScript `json:"startup_scripts"`
}

func (c *Client) CreateStartupScript(ctx context.Context, req *CreateStartupScriptRequest) (*StartupScript, error) {
	res, err := c.do(ctx, http.MethodPost, "v2/startup-scripts", &req)
	if err != nil {
		return nil, fmt.Errorf("error creating startup script: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("error creating startup script: status: %d %s msg: %s", res.StatusCode, http.StatusText(res.StatusCode), string(body))
	}

	resp := ScriptResponse{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("error decoding request: %w", err)
	}

	return &resp.Script, nil
}

func (c *Client) ListStartupScripts(ctx context.Context) (*ScriptsResponse, error) {
	res, err := c.do(ctx, http.MethodGet, "v2/startup-scripts", nil)
	if err != nil {
		return nil, fmt.Errorf("error listing startup script: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("error listing startup script: status: %d %s msg: %s", res.StatusCode, http.StatusText(res.StatusCode), string(body))
	}

	resp := ScriptsResponse{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("error decoding request: %w", err)
	}

	return &resp, nil
}

func (c *Client) GetStartupScript(ctx context.Context, id StartupScriptID) (*StartupScript, error) {
	res, err := c.do(ctx, http.MethodGet, fmt.Sprintf("v2/startup-scripts/%s", id), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting startup script: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("error getting startup script: status: %d %s msg: %s", res.StatusCode, http.StatusText(res.StatusCode), string(body))
	}

	resp := ScriptResponse{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("error decoding request: %w", err)
	}

	return &resp.Script, nil
}
