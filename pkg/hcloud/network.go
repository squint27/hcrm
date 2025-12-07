package hcloud

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// NetworkManager provides high-level operations for managing networks
type NetworkManager struct {
	client Client
}

// NewNetworkManager creates a new NetworkManager with the given client
func NewNetworkManager(client Client) *NetworkManager {
	return &NetworkManager{
		client: client,
	}
}

// GetNetworkById retrieves a network by ID
func (nm *NetworkManager) GetNetworkById(ctx context.Context, id int64) (*hcloud.Network, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid network ID: %d", id)
	}
	return nm.client.GetNetworkById(ctx, id)
}

// GetNetworkByName retrieves a network by name
func (nm *NetworkManager) GetNetworkByName(ctx context.Context, name string) (*hcloud.Network, error) {
	if name == "" {
		return nil, fmt.Errorf("network name is required")
	}
	return nm.client.GetNetworkByName(ctx, name)
}

// CreateNetwork creates a new network with the given options
func (nm *NetworkManager) CreateNetwork(ctx context.Context, opts hcloud.NetworkCreateOpts) (*hcloud.Network, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("network name is required")
	}
	return nm.client.CreateNetwork(ctx, opts)
}

// UpdateNetwork updates an existing network
func (nm *NetworkManager) UpdateNetwork(ctx context.Context, id int64, opts hcloud.NetworkUpdateOpts) (*hcloud.Network, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid network ID: %d", id)
	}
	return nm.client.UpdateNetwork(ctx, id, opts)
}

// DeleteNetwork deletes a network by ID
func (nm *NetworkManager) DeleteNetwork(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid network ID: %d", id)
	}
	return nm.client.DeleteNetwork(ctx, id)
}

// ListNetworks lists all networks
func (nm *NetworkManager) ListNetworks(ctx context.Context) ([]*hcloud.Network, error) {
	opts := hcloud.NetworkListOpts{}
	return nm.client.ListNetworks(ctx, opts)
}
