package hcloud

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// Client is an interface for interacting with Hetzner Cloud API
type Client interface {
	// Network operations
	GetNetworkById(ctx context.Context, id int64) (*hcloud.Network, error)
	GetNetworkByName(ctx context.Context, name string) (*hcloud.Network, error)
	CreateNetwork(ctx context.Context, opts hcloud.NetworkCreateOpts) (*hcloud.Network, error)
	UpdateNetwork(ctx context.Context, id int64, opts hcloud.NetworkUpdateOpts) (*hcloud.Network, error)
	DeleteNetwork(ctx context.Context, id int64) error
	ListNetworks(ctx context.Context, opts hcloud.NetworkListOpts) ([]*hcloud.Network, error)
}

// ClientImpl is the concrete implementation of the Client interface
type ClientImpl struct {
	hcloudClient *hcloud.Client
}

// NewClient creates a new HCloud client with the provided token
func NewClient(token string) (*ClientImpl, error) {
	hcloudClient := hcloud.NewClient(hcloud.WithToken(token))
	return &ClientImpl{
		hcloudClient: hcloudClient,
	}, nil
}

// NewClientWithImpl allows for dependency injection of a custom hcloud client
// This is useful for testing and mocking
func NewClientWithImpl(hcloudClient *hcloud.Client) *ClientImpl {
	return &ClientImpl{
		hcloudClient: hcloudClient,
	}
}

// GetNetwork retrieves a network by ID from Hetzner Cloud
func (c *ClientImpl) GetNetworkById(ctx context.Context, id int64) (*hcloud.Network, error) {
	network, _, err := c.hcloudClient.Network.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return network, nil
}

// GetNetworkByName retrieves a network by name from Hetzner Cloud
func (c *ClientImpl) GetNetworkByName(ctx context.Context, name string) (*hcloud.Network, error) {
	network, _, err := c.hcloudClient.Network.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return network, nil
}

// CreateNetwork creates a new network in Hetzner Cloud
func (c *ClientImpl) CreateNetwork(ctx context.Context, opts hcloud.NetworkCreateOpts) (*hcloud.Network, error) {
	network, _, err := c.hcloudClient.Network.Create(ctx, opts)
	if err != nil {
		return nil, err
	}
	return network, nil
}

// UpdateNetwork updates an existing network in Hetzner Cloud
func (c *ClientImpl) UpdateNetwork(ctx context.Context, id int64, opts hcloud.NetworkUpdateOpts) (*hcloud.Network, error) {
	network, _, err := c.hcloudClient.Network.Update(ctx, &hcloud.Network{ID: id}, opts)
	if err != nil {
		return nil, err
	}
	return network, nil
}

// DeleteNetwork deletes a network from Hetzner Cloud
func (c *ClientImpl) DeleteNetwork(ctx context.Context, id int64) error {
	_, err := c.hcloudClient.Network.Delete(ctx, &hcloud.Network{ID: id})
	return err
}

// ListNetworks lists all networks or filters them based on options
func (c *ClientImpl) ListNetworks(ctx context.Context, opts hcloud.NetworkListOpts) ([]*hcloud.Network, error) {
	networks, err := c.hcloudClient.Network.All(ctx)
	if err != nil {
		return nil, err
	}
	return networks, nil
}
