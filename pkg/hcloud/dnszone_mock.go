package hcloud

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// MockDnsZoneClient is a mock implementation of the Client interface for testing
type MockDnsZoneClient struct {
	GetZoneByIdFunc   func(ctx context.Context, id int64) (*hcloud.Zone, *hcloud.Response, error)
	GetZoneByNameFunc func(ctx context.Context, name string) (*hcloud.Zone, *hcloud.Response, error)
	CreateZoneFunc    func(ctx context.Context, name string, mode string, ttl *int, labels map[string]string) (*hcloud.Zone, *hcloud.Response, error)
	DeleteZoneFunc    func(ctx context.Context, dnszone *hcloud.Zone) (*hcloud.Response, error)
	ListZonesFunc     func(ctx context.Context) ([]*hcloud.Zone, error)
}

func (m *MockDnsZoneClient) GetZoneById(ctx context.Context, id int64) (*hcloud.Zone, *hcloud.Response, error) {
	return m.GetZoneByIdFunc(ctx, id)
}

func (m *MockDnsZoneClient) GetZoneByName(ctx context.Context, name string) (*hcloud.Zone, *hcloud.Response, error) {
	return m.GetZoneByNameFunc(ctx, name)
}

func (m *MockDnsZoneClient) CreateZone(ctx context.Context, name string, mode string, ttl *int, labels map[string]string) (*hcloud.Zone, *hcloud.Response, error) {
	return m.CreateZoneFunc(ctx, name, mode, ttl, labels)
}

func (m *MockDnsZoneClient) DeleteZone(ctx context.Context, dnszone *hcloud.Zone) (*hcloud.Response, error) {
	return m.DeleteZoneFunc(ctx, dnszone)
}

func (m *MockDnsZoneClient) ListZones(ctx context.Context) ([]*hcloud.Zone, error) {
	return m.ListZonesFunc(ctx)
}
