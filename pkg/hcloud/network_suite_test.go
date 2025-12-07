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
	var nm *NetworkManager

	BeforeEach(func() {
		mockClient = &MockClient{}
		nm = NewNetworkManager(mockClient)
	})

	Describe("GetNetworkById", func() {
		When("network exists", func() {
			BeforeEach(func() {
				mockClient.GetNetworkByIdFunc = func(ctx context.Context, id int64) (*hcloud.Network, error) {
					return &hcloud.Network{ID: 123, Name: "test-network"}, nil
				}
			})

			It("should retrieve network by ID", func() {
				network, err := nm.GetNetworkById(context.Background(), 123)
				Expect(err).NotTo(HaveOccurred())
				Expect(network).NotTo(BeNil())
				Expect(network.ID).To(Equal(int64(123)))
				Expect(network.Name).To(Equal("test-network"))
			})
		})

		When("ID is invalid", func() {
			It("should return error for zero ID", func() {
				_, err := nm.GetNetworkById(context.Background(), 0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid network ID"))
			})

			It("should return error for negative ID", func() {
				_, err := nm.GetNetworkById(context.Background(), -1)
				Expect(err).To(HaveOccurred())
			})
		})

		When("network is not found", func() {
			BeforeEach(func() {
				mockClient.GetNetworkByIdFunc = func(ctx context.Context, id int64) (*hcloud.Network, error) {
					return nil, errors.New("network not found")
				}
			})

			It("should return error", func() {
				_, err := nm.GetNetworkById(context.Background(), 999)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("network not found"))
			})
		})
	})

	Describe("CreateNetwork", func() {
		When("valid network options are provided", func() {
			BeforeEach(func() {
				mockClient.CreateNetworkFunc = func(ctx context.Context, opts hcloud.NetworkCreateOpts) (*hcloud.Network, error) {
					return &hcloud.Network{ID: 1, Name: opts.Name}, nil
				}
			})

			It("should create a network", func() {
				opts := hcloud.NetworkCreateOpts{Name: "test-network"}
				network, err := nm.CreateNetwork(context.Background(), opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(network).NotTo(BeNil())
				Expect(network.ID).To(Equal(int64(1)))
				Expect(network.Name).To(Equal("test-network"))
			})
		})

		When("name is missing", func() {
			It("should return validation error", func() {
				opts := hcloud.NetworkCreateOpts{Name: ""}
				_, err := nm.CreateNetwork(context.Background(), opts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("network name is required"))
			})
		})

		When("API returns an error", func() {
			BeforeEach(func() {
				mockClient.CreateNetworkFunc = func(ctx context.Context, opts hcloud.NetworkCreateOpts) (*hcloud.Network, error) {
					return nil, errors.New("api error")
				}
			})

			It("should propagate the error", func() {
				opts := hcloud.NetworkCreateOpts{Name: "test-network"}
				_, err := nm.CreateNetwork(context.Background(), opts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("api error"))
			})
		})
	})

	Describe("UpdateNetwork", func() {
		When("valid update options are provided", func() {
			BeforeEach(func() {
				mockClient.UpdateNetworkFunc = func(ctx context.Context, id int64, opts hcloud.NetworkUpdateOpts) (*hcloud.Network, error) {
					return &hcloud.Network{ID: id, Name: opts.Name}, nil
				}
			})

			It("should update network name", func() {
				opts := hcloud.NetworkUpdateOpts{Name: "updated-network"}
				network, err := nm.UpdateNetwork(context.Background(), 123, opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(network.Name).To(Equal("updated-network"))
			})
		})

		When("ID is invalid", func() {
			It("should return error for zero ID", func() {
				opts := hcloud.NetworkUpdateOpts{Name: "updated"}
				_, err := nm.UpdateNetwork(context.Background(), 0, opts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid network ID"))
			})
		})

		When("API returns an error", func() {
			BeforeEach(func() {
				mockClient.UpdateNetworkFunc = func(ctx context.Context, id int64, opts hcloud.NetworkUpdateOpts) (*hcloud.Network, error) {
					return nil, errors.New("update failed")
				}
			})

			It("should propagate the error", func() {
				opts := hcloud.NetworkUpdateOpts{Name: "updated"}
				_, err := nm.UpdateNetwork(context.Background(), 123, opts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("update failed"))
			})
		})
	})

	Describe("DeleteNetwork", func() {
		When("network exists", func() {
			BeforeEach(func() {
				mockClient.DeleteNetworkFunc = func(ctx context.Context, id int64) error {
					return nil
				}
			})

			It("should delete the network", func() {
				err := nm.DeleteNetwork(context.Background(), 123)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("ID is invalid", func() {
			It("should return error for zero ID", func() {
				err := nm.DeleteNetwork(context.Background(), 0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid network ID"))
			})
		})

		When("API returns an error", func() {
			BeforeEach(func() {
				mockClient.DeleteNetworkFunc = func(ctx context.Context, id int64) error {
					return errors.New("delete failed")
				}
			})

			It("should propagate the error", func() {
				err := nm.DeleteNetwork(context.Background(), 123)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("delete failed"))
			})
		})
	})

	Describe("ListNetworks", func() {
		When("networks exist", func() {
			BeforeEach(func() {
				mockClient.ListNetworksFunc = func(ctx context.Context, opts hcloud.NetworkListOpts) ([]*hcloud.Network, error) {
					return []*hcloud.Network{
						{ID: 1, Name: "network-1"},
						{ID: 2, Name: "network-2"},
					}, nil
				}
			})

			It("should return list of networks", func() {
				networks, err := nm.ListNetworks(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(networks).To(HaveLen(2))
				Expect(networks[0].Name).To(Equal("network-1"))
				Expect(networks[1].Name).To(Equal("network-2"))
			})
		})

		When("no networks exist", func() {
			BeforeEach(func() {
				mockClient.ListNetworksFunc = func(ctx context.Context, opts hcloud.NetworkListOpts) ([]*hcloud.Network, error) {
					return []*hcloud.Network{}, nil
				}
			})

			It("should return empty list", func() {
				networks, err := nm.ListNetworks(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(networks).To(BeEmpty())
			})
		})

		When("API returns an error", func() {
			BeforeEach(func() {
				mockClient.ListNetworksFunc = func(ctx context.Context, opts hcloud.NetworkListOpts) ([]*hcloud.Network, error) {
					return nil, errors.New("list failed")
				}
			})

			It("should propagate the error", func() {
				_, err := nm.ListNetworks(context.Background())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("list failed"))
			})
		})
	})

	Describe("GetNetworkByName", func() {
		When("network exists with given name", func() {
			BeforeEach(func() {
				mockClient.GetNetworkByNameFunc = func(ctx context.Context, name string) (*hcloud.Network, error) {
					return &hcloud.Network{ID: 123, Name: name}, nil
				}
			})

			It("should retrieve network by name", func() {
				network, err := nm.GetNetworkByName(context.Background(), "test-network")
				Expect(err).NotTo(HaveOccurred())
				Expect(network).NotTo(BeNil())
				Expect(network.Name).To(Equal("test-network"))
			})
		})

		When("name is empty", func() {
			It("should return validation error", func() {
				_, err := nm.GetNetworkByName(context.Background(), "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("network name is required"))
			})
		})

		When("network is not found", func() {
			BeforeEach(func() {
				mockClient.GetNetworkByNameFunc = func(ctx context.Context, name string) (*hcloud.Network, error) {
					return nil, errors.New("network not found")
				}
			})

			It("should return error", func() {
				_, err := nm.GetNetworkByName(context.Background(), "non-existent")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("network not found"))
			})
		})

		When("API returns an error", func() {
			BeforeEach(func() {
				mockClient.GetNetworkByNameFunc = func(ctx context.Context, name string) (*hcloud.Network, error) {
					return nil, errors.New("api error")
				}
			})

			It("should propagate the error", func() {
				_, err := nm.GetNetworkByName(context.Background(), "test-network")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("api error"))
			})
		})
	})
})
