package hcloud

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// MockClient is a mock implementation of the Client interface for testing
type MockClient struct {
	GetNetworkByIdFunc   func(ctx context.Context, id int64) (*hcloud.Network, error)
	GetNetworkByNameFunc func(ctx context.Context, name string) (*hcloud.Network, error)
	CreateNetworkFunc    func(ctx context.Context, opts hcloud.NetworkCreateOpts) (*hcloud.Network, error)
	UpdateNetworkFunc    func(ctx context.Context, id int64, opts hcloud.NetworkUpdateOpts) (*hcloud.Network, error)
	DeleteNetworkFunc    func(ctx context.Context, id int64) error
	ListNetworksFunc     func(ctx context.Context, opts hcloud.NetworkListOpts) ([]*hcloud.Network, error)
}

// GetNetworkById calls the mocked GetNetworkFunc
func (m *MockClient) GetNetworkById(ctx context.Context, id int64) (*hcloud.Network, error) {
	if m.GetNetworkByIdFunc != nil {
		return m.GetNetworkByIdFunc(ctx, id)
	}
	return nil, nil
}

// GetNetworkByName calls the mocked GetNetworkByNameFunc
func (m *MockClient) GetNetworkByName(ctx context.Context, name string) (*hcloud.Network, error) {
	if m.GetNetworkByNameFunc != nil {
		return m.GetNetworkByNameFunc(ctx, name)
	}
	return nil, nil
}

// CreateNetwork calls the mocked CreateNetworkFunc
func (m *MockClient) CreateNetwork(ctx context.Context, opts hcloud.NetworkCreateOpts) (*hcloud.Network, error) {
	if m.CreateNetworkFunc != nil {
		return m.CreateNetworkFunc(ctx, opts)
	}
	return nil, nil
}

// UpdateNetwork calls the mocked UpdateNetworkFunc
func (m *MockClient) UpdateNetwork(ctx context.Context, id int64, opts hcloud.NetworkUpdateOpts) (*hcloud.Network, error) {
	if m.UpdateNetworkFunc != nil {
		return m.UpdateNetworkFunc(ctx, id, opts)
	}
	return nil, nil
}

// DeleteNetwork calls the mocked DeleteNetworkFunc
func (m *MockClient) DeleteNetwork(ctx context.Context, id int64) error {
	if m.DeleteNetworkFunc != nil {
		return m.DeleteNetworkFunc(ctx, id)
	}
	return nil
}

// ListNetworks calls the mocked ListNetworksFunc
func (m *MockClient) ListNetworks(ctx context.Context, opts hcloud.NetworkListOpts) ([]*hcloud.Network, error) {
	if m.ListNetworksFunc != nil {
		return m.ListNetworksFunc(ctx, opts)
	}
	return nil, nil
}
