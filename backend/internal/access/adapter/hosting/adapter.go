package hosting

import (
	"context"
	"fmt"
	"net/url"

	"vpainless/internal/access/core"
	hosting "vpainless/internal/hosting/core"
)

type groupService interface {
	SaveGroup(ctx context.Context, g *hosting.Group) error
}

type Adapter struct {
	service groupService
}

func NewAdapter(service groupService) *Adapter {
	return &Adapter{service: service}
}

func (a *Adapter) NotifyGroupCreated(ctx context.Context, g *core.Group) error {
	url, err := url.Parse(g.Host)
	if err != nil {
		return fmt.Errorf("error parsing url %s: %w", g.Host, err)
	}

	return a.service.SaveGroup(ctx, &hosting.Group{
		ID:   hosting.GroupID{UUID: g.ID.UUID},
		Name: g.Name,
		Host: hosting.Provider{
			Base:   *url,
			Name:   hosting.Vultr,
			APIKey: g.APIKey,
		},
	})
}
