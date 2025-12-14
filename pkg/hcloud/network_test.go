package hcloud

import (
	"context"
	"errors"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NetworkManager", func() {
	var mockClient *MockClient
	var nc NetworkClient

	BeforeEach(func() {
		mockClient = &MockClient{}
		nc = NetworkClient(mockClient)
	})

	Describe("GetNetworkById", func() {
		When("network exists", func() {
			BeforeEach(func() {
				mockClient.GetNetworkByIdFunc = func(ctx context.Context, id int64) (*hcloud.Network, *hcloud.Response, error) {
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
				mockClient.GetNetworkByIdFunc = func(ctx context.Context, id int64) (*hcloud.Network, *hcloud.Response, error) {
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
				mockClient.CreateNetworkFunc = func(ctx context.Context, name string, ipRange string) (*hcloud.Network, *hcloud.Response, error) {
					return &hcloud.Network{ID: 1, Name: "created-network"}, nil, nil
				}
			})

			It("should create a network", func() {
				network, _, err := nc.CreateNetwork(context.Background(), "created-network", "192.168.0.0/24")
				Expect(err).NotTo(HaveOccurred())
				Expect(network).NotTo(BeNil())
				Expect(network.ID).To(Equal(int64(1)))
				Expect(network.Name).To(Equal("created-network"))
			})
		})

		When("API returns an error", func() {
			BeforeEach(func() {
				mockClient.CreateNetworkFunc = func(ctx context.Context, name string, ipRange string) (*hcloud.Network, *hcloud.Response, error) {
					return nil, nil, errors.New("api error")
				}
			})

			It("should propagate the error", func() {
				_, _, err := nc.CreateNetwork(context.Background(), "created-network", "192.168.0.0/24")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("api error"))
			})
		})
	})

	Describe("UpdateNetwork", func() {
		When("valid update options are provided", func() {
			BeforeEach(func() {
				mockClient.UpdateNetworkFunc = func(ctx context.Context, network *hcloud.Network, opts hcloud.NetworkUpdateOpts) (*hcloud.Network, *hcloud.Response, error) {
					return &hcloud.Network{ID: network.ID, Name: opts.Name}, nil, nil
				}
			})
			var network = &hcloud.Network{ID: 123, Name: "original-network"}
			It("should update network name", func() {
				opts := hcloud.NetworkUpdateOpts{Name: "updated-network"}
				updatedNetwork, _, err := nc.UpdateNetwork(context.Background(), network, opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedNetwork.Name).To(Equal("updated-network"))
			})
		})

		When("API returns an error", func() {
			BeforeEach(func() {
				mockClient.UpdateNetworkFunc = func(ctx context.Context, network *hcloud.Network, opts hcloud.NetworkUpdateOpts) (*hcloud.Network, *hcloud.Response, error) {
					return nil, nil, errors.New("update failed")
				}
			})

			var network = &hcloud.Network{ID: 123, Name: "original-network"}
			It("should propagate the error", func() {
				opts := hcloud.NetworkUpdateOpts{Name: "updated"}
				_, _, err := nc.UpdateNetwork(context.Background(), network, opts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("update failed"))
			})
		})
	})

	Describe("DeleteNetwork", func() {
		When("network exists", func() {
			BeforeEach(func() {
				mockClient.DeleteNetworkFunc = func(ctx context.Context, network *hcloud.Network) (*hcloud.Response, error) {
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
				mockClient.DeleteNetworkFunc = func(ctx context.Context, network *hcloud.Network) (*hcloud.Response, error) {
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
				mockClient.ListNetworksFunc = func(ctx context.Context) ([]*hcloud.Network, error) {
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
				mockClient.ListNetworksFunc = func(ctx context.Context) ([]*hcloud.Network, error) {
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
				mockClient.ListNetworksFunc = func(ctx context.Context) ([]*hcloud.Network, error) {
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
				mockClient.GetNetworkByNameFunc = func(ctx context.Context, name string) (*hcloud.Network, *hcloud.Response, error) {
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
				mockClient.GetNetworkByNameFunc = func(ctx context.Context, name string) (*hcloud.Network, *hcloud.Response, error) {
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
				mockClient.GetNetworkByNameFunc = func(ctx context.Context, name string) (*hcloud.Network, *hcloud.Response, error) {
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
