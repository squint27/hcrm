package hcloud

import (
	"context"
	"net"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// Client is an interface for interacting with Hetzner Cloud API
type NetworkClient interface {
	// Network operations
	GetNetworkById(ctx context.Context, id int64) (*hcloud.Network, *hcloud.Response, error)
	GetNetworkByName(ctx context.Context, name string) (*hcloud.Network, *hcloud.Response, error)
	CreateNetwork(ctx context.Context, name string, ipRange string) (*hcloud.Network, *hcloud.Response, error)
	UpdateNetwork(ctx context.Context, network *hcloud.Network, opts hcloud.NetworkUpdateOpts) (*hcloud.Network, *hcloud.Response, error)
	DeleteNetwork(ctx context.Context, network *hcloud.Network) (*hcloud.Response, error)
	ListNetworks(ctx context.Context) ([]*hcloud.Network, error)
}

type hcloudNetworkAdapter struct {
	client *hcloud.Client
}

// NewClient creates a new HCloud client with the provided token
func NewNetworkClient(token string) *hcloudNetworkAdapter {
	client := hcloud.NewClient(hcloud.WithToken(token))
	return &hcloudNetworkAdapter{
		client: client,
	}
}

// GetNetworkById retrieves a network by ID
func (a *hcloudNetworkAdapter) GetNetworkById(ctx context.Context, id int64) (*hcloud.Network, *hcloud.Response, error) {
	return a.client.Network.GetByID(ctx, id)
}

// GetNetworkByName retrieves a network by name
func (a *hcloudNetworkAdapter) GetNetworkByName(ctx context.Context, name string) (*hcloud.Network, *hcloud.Response, error) {
	return a.client.Network.GetByName(ctx, name)
}

// CreateNetwork creates a new network with the given options
func (a *hcloudNetworkAdapter) CreateNetwork(ctx context.Context, name string, ipRange string) (*hcloud.Network, *hcloud.Response, error) {
	_, cidr, err := net.ParseCIDR(ipRange)
	if err != nil {
		return nil, nil, err
	}
	opts := hcloud.NetworkCreateOpts{
		Name:    name,
		IPRange: cidr,
	}
	return a.client.Network.Create(ctx, opts)
}

// UpdateNetwork updates an existing network
func (a *hcloudNetworkAdapter) UpdateNetwork(ctx context.Context, network *hcloud.Network, opts hcloud.NetworkUpdateOpts) (*hcloud.Network, *hcloud.Response, error) {
	return a.client.Network.Update(ctx, network, opts)
}

// DeleteNetwork deletes a network
func (a *hcloudNetworkAdapter) DeleteNetwork(ctx context.Context, network *hcloud.Network) (*hcloud.Response, error) {
	return a.client.Network.Delete(ctx, network)
}

// ListNetworks lists all networks
func (a *hcloudNetworkAdapter) ListNetworks(ctx context.Context) ([]*hcloud.Network, error) {
	return a.client.Network.All(ctx)
}
