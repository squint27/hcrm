package hcloud

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// MockClient is a mock implementation of the Client interface for testing
type MockClient struct {
	GetNetworkByIdFunc   func(ctx context.Context, id int64) (*hcloud.Network, *hcloud.Response, error)
	GetNetworkByNameFunc func(ctx context.Context, name string) (*hcloud.Network, *hcloud.Response, error)
	CreateNetworkFunc    func(ctx context.Context, opts hcloud.NetworkCreateOpts) (*hcloud.Network, *hcloud.Response, error)
	UpdateNetworkFunc    func(ctx context.Context, network *hcloud.Network, opts hcloud.NetworkUpdateOpts) (*hcloud.Network, *hcloud.Response, error)
	DeleteNetworkFunc    func(ctx context.Context, network *hcloud.Network) (*hcloud.Response, error)
	ListNetworksFunc     func(ctx context.Context) ([]*hcloud.Network, error)
}

// GetNetworkById calls the mocked GetNetworkFunc
func (m *MockClient) GetNetworkById(ctx context.Context, id int64) (*hcloud.Network, *hcloud.Response, error) {
	if m.GetNetworkByIdFunc != nil {
		return m.GetNetworkByIdFunc(ctx, id)
	}
	return nil, nil, nil
}

// GetNetworkByName calls the mocked GetNetworkByNameFunc
func (m *MockClient) GetNetworkByName(ctx context.Context, name string) (*hcloud.Network, *hcloud.Response, error) {
	if m.GetNetworkByNameFunc != nil {
		return m.GetNetworkByNameFunc(ctx, name)
	}
	return nil, nil, nil
}

// CreateNetwork calls the mocked CreateNetworkFunc
func (m *MockClient) CreateNetwork(ctx context.Context, opts hcloud.NetworkCreateOpts) (*hcloud.Network, *hcloud.Response, error) {
	if m.CreateNetworkFunc != nil {
		return m.CreateNetworkFunc(ctx, opts)
	}
	return nil, nil, nil
}

// UpdateNetwork calls the mocked UpdateNetworkFunc
func (m *MockClient) UpdateNetwork(ctx context.Context, network *hcloud.Network, opts hcloud.NetworkUpdateOpts) (*hcloud.Network, *hcloud.Response, error) {
	if m.UpdateNetworkFunc != nil {
		return m.UpdateNetworkFunc(ctx, network, opts)
	}
	return nil, nil, nil
}

// DeleteNetwork calls the mocked DeleteNetworkFunc
func (m *MockClient) DeleteNetwork(ctx context.Context, network *hcloud.Network) (*hcloud.Response, error) {
	if m.DeleteNetworkFunc != nil {
		return m.DeleteNetworkFunc(ctx, network)
	}
	return nil, nil
}

// ListNetworks calls the mocked ListNetworksFunc
func (m *MockClient) ListNetworks(ctx context.Context) ([]*hcloud.Network, error) {
	if m.ListNetworksFunc != nil {
		return m.ListNetworksFunc(ctx)
	}
	return nil, nil
}
