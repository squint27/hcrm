package hcloud

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// MockClient is a mock implementation of the Client interface for testing
type MockClient struct {
	GetNetworkByIdFunc      func(ctx context.Context, id int64) (*hcloud.Network, *hcloud.Response, error)
	GetNetworkByNameFunc    func(ctx context.Context, name string) (*hcloud.Network, *hcloud.Response, error)
	CreateNetworkFunc       func(ctx context.Context, name string, ipRange string, labels map[string]string) (*hcloud.Network, *hcloud.Response, error)
	UpdateNetworkLabelsFunc func(ctx context.Context, network *hcloud.Network, labels map[string]string) (*hcloud.Network, *hcloud.Response, error)
	UpdateNetworkCidrFunc   func(ctx context.Context, network *hcloud.Network, cidr string) (*hcloud.Network, *hcloud.Response, error)
	DeleteNetworkFunc       func(ctx context.Context, network *hcloud.Network) (*hcloud.Response, error)
	ListNetworksFunc        func(ctx context.Context) ([]*hcloud.Network, error)
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
func (m *MockClient) CreateNetwork(ctx context.Context, name string, ipRange string, labels map[string]string) (*hcloud.Network, *hcloud.Response, error) {
	if m.CreateNetworkFunc != nil {
		return m.CreateNetworkFunc(ctx, name, ipRange, labels)
	}
	return nil, nil, nil
}

// UpdateNetworkLabels calls the mocked UpdateNetworkLabelsFunc
func (m *MockClient) UpdateNetworkLabels(ctx context.Context, network *hcloud.Network, labels map[string]string) (*hcloud.Network, *hcloud.Response, error) {
	if m.UpdateNetworkLabelsFunc != nil {
		return m.UpdateNetworkLabelsFunc(ctx, network, labels)
	}
	return nil, nil, nil
}

// UpdateNetworkCidr calls the mocked UpdateNetworkCidrFunc
func (m *MockClient) UpdateNetworkCidr(ctx context.Context, network *hcloud.Network, cidr string) (*hcloud.Network, *hcloud.Response, error) {
	if m.UpdateNetworkCidrFunc != nil {
		return m.UpdateNetworkCidrFunc(ctx, network, cidr)
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
