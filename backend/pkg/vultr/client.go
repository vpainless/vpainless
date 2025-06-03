package vultr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

var ErrNotFound = errors.New("not found")

type ctxtype string

type Client struct {
	host url.URL
}

func NewClient(host url.URL) *Client {
	return &Client{host}
}

func (c *Client) WithAPIKey(ctx context.Context, apikey string) context.Context {
	return context.WithValue(ctx, ctxtype("apikey"), apikey)
}

func (c *Client) do(ctx context.Context, method, path string, requestBody any) (*http.Response, error) {
	apikey, ok := ctx.Value(ctxtype("apikey")).(string)
	if !ok {
		return nil, fmt.Errorf("api key missing on the context")
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.host.JoinPath(path).String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apikey))
	return http.DefaultClient.Do(request)
}
