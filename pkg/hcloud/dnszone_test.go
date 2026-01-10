package hcloud

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DnsZoneManager", func() {

	var mockDnsZoneClient *MockDnsZoneClient
	var zc DnsZoneClient

	BeforeEach(func() {
		mockDnsZoneClient = &MockDnsZoneClient{}
		zc = DnsZoneClient(mockDnsZoneClient)
	})

	Describe("GetZoneById", func() {
		When("zone exists", func() {
			BeforeEach(func() {
				mockDnsZoneClient.GetZoneByIdFunc = func(ctx context.Context, id int64) (*hcloud.Zone, *hcloud.Response, error) {
					return &hcloud.Zone{ID: 123, Name: "example.com"}, nil, nil
				}
			})

			It("should retrieve zone by ID", func() {
				zone, _, err := zc.GetZoneById(context.Background(), 123)
				Expect(err).NotTo(HaveOccurred())
				Expect(zone).NotTo(BeNil())
				Expect(zone.ID).To(Equal(int64(123)))
				Expect(zone.Name).To(Equal("example.com"))
			})
		})
	})
})
