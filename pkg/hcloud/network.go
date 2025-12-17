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
	CreateNetwork(ctx context.Context, name string, ipRange string, labels map[string]string) (*hcloud.Network, *hcloud.Response, error)
	UpdateNetworkLabels(ctx context.Context, network *hcloud.Network, labels map[string]string) (*hcloud.Network, *hcloud.Response, error)
	UpdateNetworkCidr(ctx context.Context, network *hcloud.Network, cidr string) (*hcloud.Network, *hcloud.Response, error)
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
func (a *hcloudNetworkAdapter) CreateNetwork(ctx context.Context, name string, ipRange string, labels map[string]string) (*hcloud.Network, *hcloud.Response, error) {
	_, cidr, err := net.ParseCIDR(ipRange)
	if err != nil {
		return nil, nil, err
	}
	opts := hcloud.NetworkCreateOpts{
		Name:    name,
		IPRange: cidr,
		Labels:  labels,
	}
	return a.client.Network.Create(ctx, opts)
}

// UpdateNetwork updates an existing network
func (a *hcloudNetworkAdapter) UpdateNetworkLabels(ctx context.Context, network *hcloud.Network, labels map[string]string) (*hcloud.Network, *hcloud.Response, error) {
	opts := hcloud.NetworkUpdateOpts{
		Labels: labels,
	}
	return a.client.Network.Update(ctx, network, opts)
}

// UpdateNetworkCidr updates the CIDR of an existing network
func (a *hcloudNetworkAdapter) UpdateNetworkCidr(ctx context.Context, network *hcloud.Network, cidr string) (*hcloud.Network, *hcloud.Response, error) {
	_, parsedCidr, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, nil, err
	}
	opts := hcloud.NetworkChangeIPRangeOpts{
		IPRange: parsedCidr,
	}
	action, resp, err := a.client.Network.ChangeIPRange(ctx, network, opts)
	if err != nil {
		return nil, resp, err
	}
	// Wait for the action to complete
	err = a.client.Action.WaitFor(ctx, action)
	if err != nil {
		return nil, resp, err
	}
	// Retrieve the updated network
	updatedNetwork, resp, err := a.client.Network.GetByID(ctx, network.ID)
	if err != nil {
		return nil, resp, err
	}
	return updatedNetwork, resp, nil
}

// DeleteNetwork deletes a network
func (a *hcloudNetworkAdapter) DeleteNetwork(ctx context.Context, network *hcloud.Network) (*hcloud.Response, error) {
	return a.client.Network.Delete(ctx, network)
}

// ListNetworks lists all networks
func (a *hcloudNetworkAdapter) ListNetworks(ctx context.Context) ([]*hcloud.Network, error) {
	return a.client.Network.All(ctx)
}
