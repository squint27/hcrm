package hcloud

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// Client is an interface for interacting with Hetzner Cloud API
type DnsZoneClient interface {
	GetZoneById(ctx context.Context, id int64) (*hcloud.Zone, *hcloud.Response, error)
	GetZoneByName(ctx context.Context, name string) (*hcloud.Zone, *hcloud.Response, error)
	CreateZone(ctx context.Context, name string, mode string, ttl *int, labels map[string]string) (*hcloud.Zone, *hcloud.Response, error)
	// UpdateZone(ctx context.Context, zone *hcloud.Zone, name string, labels map[string]string) (*hcloud.Zone, *hcloud.Response, error)
	DeleteZone(ctx context.Context, zone *hcloud.Zone) (*hcloud.Response, error)
	ListZones(ctx context.Context) ([]*hcloud.Zone, error)
}

type hcloudDnsZoneAdapter struct {
	client *hcloud.Client
}

// NewClient creates a new HCloud client with the provided token
func NewDnsZoneClient(token string) *hcloudDnsZoneAdapter {
	client := hcloud.NewClient(hcloud.WithToken(token))
	return &hcloudDnsZoneAdapter{
		client: client,
	}
}

func (a *hcloudDnsZoneAdapter) GetZoneById(ctx context.Context, id int64) (*hcloud.Zone, *hcloud.Response, error) {
	return a.client.Zone.GetByID(ctx, id)
}

func (a *hcloudDnsZoneAdapter) GetZoneByName(ctx context.Context, name string) (*hcloud.Zone, *hcloud.Response, error) {
	return a.client.Zone.GetByName(ctx, name)
}

func (a *hcloudDnsZoneAdapter) CreateZone(ctx context.Context, name string, mode string, ttl *int, labels map[string]string) (*hcloud.Zone, *hcloud.Response, error) {
	opts := hcloud.ZoneCreateOpts{
		Name:   name,
		Mode:   hcloud.ZoneMode(mode),
		TTL:    ttl,
		Labels: labels,
	}

	result, response, err := a.client.Zone.Create(ctx, opts)
	if err != nil {
		return nil, response, err
	}

	err = a.client.Action.WaitFor(ctx, result.Action)
	if err != nil {
		zone := result.Zone
		return zone, response, nil
	} else {
		return nil, response, err
	}

}

func (a *hcloudDnsZoneAdapter) DeleteZone(ctx context.Context, zone *hcloud.Zone) (*hcloud.Response, error) {

	result, response, err := a.client.Zone.Delete(ctx, zone)
	if err != nil {
		return response, err
	}

	err = a.client.Action.WaitFor(ctx, result.Action)

	return response, err
}

func (a *hcloudDnsZoneAdapter) ListZones(ctx context.Context) ([]*hcloud.Zone, error) {
	zones, err := a.client.Zone.All(ctx)
	if err != nil {
		return nil, err
	}
	return zones, nil
}
