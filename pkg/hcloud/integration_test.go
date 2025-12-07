//go:build integration
// +build integration

package hcloud

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HCloud Integration Tests", func() {
	var client *ClientImpl
	var nm *NetworkManager
	var token string

	BeforeSuite(func() {
		token = os.Getenv("HCLOUD_TOKEN")
		if token == "" {
			Skip("HCLOUD_TOKEN not set, skipping integration tests")
		}

		var err error
		client, err = NewClient(token)
		Expect(err).NotTo(HaveOccurred())
		nm = NewNetworkManager(client)
	})

	Describe("Create and Get Network", func() {
		var testNetworkName string
		var createdNetwork *hcloud.Network

		BeforeEach(func() {
			testNetworkName = fmt.Sprintf("hcrm-test-network-%d", time.Now().Unix())
		})

		It("should create a network with IP range", func() {
			ipRangeStr := fmt.Sprintf("10.%d.0.0/24", time.Now().Unix()%250)
			_, ipnet, err := net.ParseCIDR(ipRangeStr)
			Expect(err).NotTo(HaveOccurred())

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			var createErr error
			createdNetwork, createErr = nm.CreateNetwork(ctx, hcloud.NetworkCreateOpts{
				Name:    testNetworkName,
				IPRange: ipnet,
			})
			Expect(createErr).NotTo(HaveOccurred())
			Expect(createdNetwork).NotTo(BeNil())
			Expect(createdNetwork.Name).To(Equal(testNetworkName))
		})

		It("should retrieve network by ID", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			retrieved, err := nm.GetNetworkById(ctx, createdNetwork.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrieved).NotTo(BeNil())
			Expect(retrieved.ID).To(Equal(createdNetwork.ID))
			Expect(retrieved.Name).To(Equal(testNetworkName))
		})

		It("should retrieve network by name", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			retrieved, err := nm.GetNetworkByName(ctx, testNetworkName)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrieved).NotTo(BeNil())
			Expect(retrieved.ID).To(Equal(createdNetwork.ID))
		})

		AfterEach(func() {
			if createdNetwork != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				deleteErr := nm.DeleteNetwork(ctx, createdNetwork.ID)
				Expect(deleteErr).NotTo(HaveOccurred())
			}
		})
	})

	Describe("Update Network", func() {
		var testNetworkName string
		var createdNetwork *hcloud.Network

		BeforeEach(func() {
			testNetworkName = fmt.Sprintf("hcrm-test-update-%d", time.Now().Unix())
			ipRangeStr := fmt.Sprintf("10.%d.0.0/24", time.Now().Unix()%250)
			_, ipnet, err := net.ParseCIDR(ipRangeStr)
			Expect(err).NotTo(HaveOccurred())

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			var createErr error
			createdNetwork, createErr = nm.CreateNetwork(ctx, hcloud.NetworkCreateOpts{
				Name:    testNetworkName,
				IPRange: ipnet,
			})
			Expect(createErr).NotTo(HaveOccurred())
		})

		It("should update network name and labels", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			updatedName := testNetworkName + "-updated"
			updated, err := nm.UpdateNetwork(ctx, createdNetwork.ID, hcloud.NetworkUpdateOpts{
				Name: updatedName,
				Labels: map[string]string{
					"test":    "true",
					"version": "1.0",
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(updated.Name).To(Equal(updatedName))
			Expect(updated.Labels["test"]).To(Equal("true"))
			Expect(updated.Labels["version"]).To(Equal("1.0"))
		})

		AfterEach(func() {
			if createdNetwork != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				deleteErr := nm.DeleteNetwork(ctx, createdNetwork.ID)
				Expect(deleteErr).NotTo(HaveOccurred())
			}
		})
	})

	Describe("List Networks", func() {
		It("should list all networks without error", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			networks, err := nm.ListNetworks(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(networks).NotTo(BeNil())
		})
	})

	Describe("Full Lifecycle", func() {
		var testNetworkName string
		var network *hcloud.Network

		BeforeEach(func() {
			testNetworkName = fmt.Sprintf("hcrm-test-lifecycle-%d", time.Now().Unix())
		})

		It("should create, read, update, list, and delete a network", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			// CREATE
			ipRangeStr := fmt.Sprintf("10.%d.0.0/24", time.Now().Unix()%250)
			_, ipnet, err := net.ParseCIDR(ipRangeStr)
			Expect(err).NotTo(HaveOccurred())

			var createErr error
			network, createErr = nm.CreateNetwork(ctx, hcloud.NetworkCreateOpts{
				Name:    testNetworkName,
				IPRange: ipnet,
			})
			Expect(createErr).NotTo(HaveOccurred())
			Expect(network.ID).NotTo(BeZero())

			// READ BY ID
			readByID, readErr := nm.GetNetworkById(ctx, network.ID)
			Expect(readErr).NotTo(HaveOccurred())
			Expect(readByID.ID).To(Equal(network.ID))
			Expect(readByID.Name).To(Equal(testNetworkName))

			// READ BY NAME
			readByName, readNameErr := nm.GetNetworkByName(ctx, testNetworkName)
			Expect(readNameErr).NotTo(HaveOccurred())
			Expect(readByName.ID).To(Equal(network.ID))

			// UPDATE
			newName := testNetworkName + "-modified"
			updated, updateErr := nm.UpdateNetwork(ctx, network.ID, hcloud.NetworkUpdateOpts{
				Name: newName,
				Labels: map[string]string{
					"environment": "test",
					"managed-by":  "hcrm",
				},
			})
			Expect(updateErr).NotTo(HaveOccurred())
			Expect(updated.Name).To(Equal(newName))

			// LIST
			networks, listErr := nm.ListNetworks(ctx)
			Expect(listErr).NotTo(HaveOccurred())
			found := false
			for _, n := range networks {
				if n.ID == network.ID {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())

			// DELETE
			deleteErr := nm.DeleteNetwork(ctx, network.ID)
			Expect(deleteErr).NotTo(HaveOccurred())
		})
	})
})
