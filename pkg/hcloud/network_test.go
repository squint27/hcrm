package hcloud

import (
	"context"
	"errors"
	"net"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NetworkManager", func() {
	var mockNetworkClient *MockNetworkClient
	var nc NetworkClient

	BeforeEach(func() {
		mockNetworkClient = &MockNetworkClient{}
		nc = NetworkClient(mockNetworkClient)
	})

	Describe("GetNetworkById", func() {
		When("network exists", func() {
			BeforeEach(func() {
				mockNetworkClient.GetNetworkByIdFunc = func(ctx context.Context, id int64) (*hcloud.Network, *hcloud.Response, error) {
					return &hcloud.Network{ID: 123, Name: "test-network"}, nil, nil
				}
			})

			It("should retrieve network by ID", func() {
				network, _, err := nc.GetNetworkById(context.Background(), 123)
				Expect(err).NotTo(HaveOccurred())
				Expect(network).NotTo(BeNil())
				Expect(network.ID).To(Equal(int64(123)))
				Expect(network.Name).To(Equal("test-network"))
			})
		})

		When("network is not found", func() {
			BeforeEach(func() {
				mockNetworkClient.GetNetworkByIdFunc = func(ctx context.Context, id int64) (*hcloud.Network, *hcloud.Response, error) {
					return nil, nil, errors.New("network not found")
				}
			})

			It("should return error", func() {
				_, _, err := nc.GetNetworkById(context.Background(), 999)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("network not found"))
			})
		})
	})

	Describe("CreateNetwork", func() {
		When("valid network options are provided", func() {
			BeforeEach(func() {
				mockNetworkClient.CreateNetworkFunc = func(ctx context.Context, name string, ipRange string, labels map[string]string) (*hcloud.Network, *hcloud.Response, error) {
					return &hcloud.Network{ID: 1, Name: name, Labels: labels}, nil, nil
				}
			})

			It("should create a network", func() {
				network, _, err := nc.CreateNetwork(context.Background(), "created-network", "192.168.0.0/24", map[string]string{"key": "value"})
				Expect(err).NotTo(HaveOccurred())
				Expect(network).NotTo(BeNil())
				Expect(network.ID).To(Equal(int64(1)))
				Expect(network.Name).To(Equal("created-network"))
			})
		})

		When("API returns an error", func() {
			BeforeEach(func() {
				mockNetworkClient.CreateNetworkFunc = func(ctx context.Context, name string, ipRange string, labels map[string]string) (*hcloud.Network, *hcloud.Response, error) {
					return nil, nil, errors.New("api error")
				}
			})

			It("should propagate the error", func() {
				_, _, err := nc.CreateNetwork(context.Background(), "created-network", "192.168.0.0/24", map[string]string{"key": "value"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("api error"))
			})
		})
	})

	Describe("UpdateNetworkLabels", func() {
		When("valid update options are provided", func() {
			BeforeEach(func() {
				mockNetworkClient.UpdateNetworkLabelsFunc = func(ctx context.Context, network *hcloud.Network, labels map[string]string) (*hcloud.Network, *hcloud.Response, error) {
					return &hcloud.Network{ID: network.ID, Name: network.Name, Labels: labels}, nil, nil
				}
			})
			var network = &hcloud.Network{
				ID:   123,
				Name: "update-labels-network",
				Labels: map[string]string{
					"oldKey": "oldValue",
				},
			}
			It("should update network labels", func() {
				labels := map[string]string{"newKey": "newValue"}
				updatedNetwork, _, err := nc.UpdateNetworkLabels(context.Background(), network, labels)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedNetwork.Labels).To(Equal(labels))
			})
		})

		When("API returns an error", func() {
			BeforeEach(func() {
				mockNetworkClient.UpdateNetworkLabelsFunc = func(ctx context.Context, network *hcloud.Network, labels map[string]string) (*hcloud.Network, *hcloud.Response, error) {
					return nil, nil, errors.New("update failed")
				}
			})

			var network = &hcloud.Network{
				ID:   123,
				Name: "update-labels-fail-network",
				Labels: map[string]string{
					"oldKey": "oldValue",
				},
			}
			It("should propagate the error", func() {
				labels := map[string]string{"newKey": "newValue"}
				_, _, err := nc.UpdateNetworkLabels(context.Background(), network, labels)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("update failed"))
			})
		})
	})

	Describe("UpdateNetworkCidr", func() {
		When("valid update options are provided", func() {
			BeforeEach(func() {
				_, expectedCidr, err := net.ParseCIDR("10.0.0.0/8")
				Expect(err).NotTo(HaveOccurred())

				mockNetworkClient.UpdateNetworkCidrFunc = func(ctx context.Context, network *hcloud.Network, cidr string) (*hcloud.Network, *hcloud.Response, error) {
					return &hcloud.Network{
						ID:      network.ID,
						Name:    network.Name,
						IPRange: expectedCidr,
					}, nil, nil
				}
			})
			_, oldIpRange, err := net.ParseCIDR("10.0.0.0/16")
			Expect(err).NotTo(HaveOccurred())
			_, newIpRange, err := net.ParseCIDR("10.0.0.0/8")
			Expect(err).NotTo(HaveOccurred())

			var network = &hcloud.Network{
				ID:      123,
				Name:    "update-cidr-network",
				IPRange: oldIpRange,
			}
			It("should update network cidr", func() {
				updatedNetwork, _, err := nc.UpdateNetworkCidr(context.Background(), network, newIpRange.String())
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedNetwork.IPRange).To(Equal(newIpRange))
			})
		})

		When("API returns an error", func() {
			BeforeEach(func() {
				mockNetworkClient.UpdateNetworkCidrFunc = func(ctx context.Context, network *hcloud.Network, cidr string) (*hcloud.Network, *hcloud.Response, error) {
					return nil, nil, errors.New("update failed")
				}
			})

			_, oldIpRange, err := net.ParseCIDR("10.0.0.0/16")
			Expect(err).NotTo(HaveOccurred())
			_, newIpRange, err := net.ParseCIDR("10.0.0.0/8")
			Expect(err).NotTo(HaveOccurred())

			var network = &hcloud.Network{
				ID:      123,
				Name:    "update-labels-network",
				IPRange: oldIpRange,
			}
			It("should propagate the error", func() {
				_, _, err := nc.UpdateNetworkCidr(context.Background(), network, newIpRange.String())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("update failed"))
			})
		})
	})

	Describe("DeleteNetwork", func() {
		When("network exists", func() {
			BeforeEach(func() {
				mockNetworkClient.DeleteNetworkFunc = func(ctx context.Context, network *hcloud.Network) (*hcloud.Response, error) {
					return nil, nil
				}
			})

			var network = &hcloud.Network{ID: 123, Name: "to-be-deleted"}
			It("should delete the network", func() {
				_, err := nc.DeleteNetwork(context.Background(), network)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("API returns an error", func() {
			BeforeEach(func() {
				mockNetworkClient.DeleteNetworkFunc = func(ctx context.Context, network *hcloud.Network) (*hcloud.Response, error) {
					return nil, errors.New("delete failed")
				}
			})

			var network = &hcloud.Network{ID: 123, Name: "to-be-deleted"}
			It("should propagate the error", func() {
				_, err := nc.DeleteNetwork(context.Background(), network)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("delete failed"))
			})
		})
	})

	Describe("ListNetworks", func() {
		When("networks exist", func() {
			BeforeEach(func() {
				mockNetworkClient.ListNetworksFunc = func(ctx context.Context) ([]*hcloud.Network, error) {
					return []*hcloud.Network{
						{ID: 1, Name: "network-1"},
						{ID: 2, Name: "network-2"},
					}, nil
				}
			})

			It("should return list of networks", func() {
				networks, err := nc.ListNetworks(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(networks).To(HaveLen(2))
				Expect(networks[0].Name).To(Equal("network-1"))
				Expect(networks[1].Name).To(Equal("network-2"))
			})
		})

		When("no networks exist", func() {
			BeforeEach(func() {
				mockNetworkClient.ListNetworksFunc = func(ctx context.Context) ([]*hcloud.Network, error) {
					return []*hcloud.Network{}, nil
				}
			})

			It("should return empty list", func() {
				networks, err := nc.ListNetworks(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(networks).To(BeEmpty())
			})
		})

		When("API returns an error", func() {
			BeforeEach(func() {
				mockNetworkClient.ListNetworksFunc = func(ctx context.Context) ([]*hcloud.Network, error) {
					return nil, errors.New("list failed")
				}
			})

			It("should propagate the error", func() {
				_, err := nc.ListNetworks(context.Background())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("list failed"))
			})
		})
	})

	Describe("GetNetworkByName", func() {
		When("network exists with given name", func() {
			BeforeEach(func() {
				mockNetworkClient.GetNetworkByNameFunc = func(ctx context.Context, name string) (*hcloud.Network, *hcloud.Response, error) {
					return &hcloud.Network{ID: 123, Name: name}, nil, nil
				}
			})

			It("should retrieve network by name", func() {
				network, _, err := nc.GetNetworkByName(context.Background(), "test-network")
				Expect(err).NotTo(HaveOccurred())
				Expect(network).NotTo(BeNil())
				Expect(network.Name).To(Equal("test-network"))
			})
		})

		When("network is not found", func() {
			BeforeEach(func() {
				mockNetworkClient.GetNetworkByNameFunc = func(ctx context.Context, name string) (*hcloud.Network, *hcloud.Response, error) {
					return nil, nil, errors.New("network not found")
				}
			})

			It("should return error", func() {
				_, _, err := nc.GetNetworkByName(context.Background(), "non-existent")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("network not found"))
			})
		})

		When("API returns an error", func() {
			BeforeEach(func() {
				mockNetworkClient.GetNetworkByNameFunc = func(ctx context.Context, name string) (*hcloud.Network, *hcloud.Response, error) {
					return nil, nil, errors.New("api error")
				}
			})

			It("should propagate the error", func() {
				_, _, err := nc.GetNetworkByName(context.Background(), "test-network")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("api error"))
			})
		})
	})
})
