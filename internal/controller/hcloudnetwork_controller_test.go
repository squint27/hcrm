/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	hcloudv1alpha1 "bunskin.com/hcrm/api/v1alpha1"
	"bunskin.com/hcrm/pkg/hcloud"
	hcloudgo "github.com/hetznercloud/hcloud-go/v2/hcloud"
)

var _ = Describe("HcloudNetwork Controller", func() {
	Context("Create new HcloudNetwork", func() {
		const namespace = "default"

		ctx := context.Background()

		It("should successfully create a network in Hetzner Cloud", func() {
			const resourceName = "test-create-success"
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: namespace,
			}

			By("creating the HcloudNetwork resource")
			resource := &hcloudv1alpha1.HcloudNetwork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: hcloudv1alpha1.HcloudNetworkSpec{
					Name:    "test-network-create",
					IpRange: "10.0.0.0/8",
					Labels: map[string]string{
						"env": "test",
					},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			By("creating a mock HCloud manager")
			mockMgr := &hcloud.MockClient{}
			createdNetwork := &hcloudgo.Network{
				ID:   12345,
				Name: "test-network-create",
				IPRange: &net.IPNet{
					IP:   net.IPv4(10, 0, 0, 0),
					Mask: net.IPv4Mask(255, 0, 0, 0),
				},
				Labels: map[string]string{"env": "test"},
			}

			mockMgr.CreateNetworkFunc = func(ctx context.Context, opts hcloudgo.NetworkCreateOpts) (*hcloudgo.Network, error) {
				return createdNetwork, nil
			}

			manager := hcloud.NewNetworkManager(mockMgr)

			By("reconciling the resource")
			reconciler := &HcloudNetworkReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				HCloudMgr: manager,
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("verifying the resource status was updated")
			updatedResource := &hcloudv1alpha1.HcloudNetwork{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updatedResource)).To(Succeed())
			Expect(updatedResource.Status.NetworkId).To(Equal(12345))
			Expect(updatedResource.Status.ObservedGeneration).To(Equal(updatedResource.Generation))

			By("verifying the Available condition was set")
			condition := meta.FindStatusCondition(updatedResource.Status.Conditions, "Available")
			Expect(condition).NotTo(BeNil())
			Expect(condition.Status).To(Equal(metav1.ConditionTrue))
			Expect(condition.Reason).To(Equal("NetworkCreated"))

			By("verifying finalizer was added")
			Expect(updatedResource.ObjectMeta.Finalizers).To(ContainElement(finalizerName))

			By("cleaning up the resource")
			Expect(k8sClient.Delete(ctx, updatedResource)).To(Succeed())
		})

		It("should handle invalid IP range gracefully", func() {
			const resourceName = "test-invalid-cidr"
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: namespace,
			}

			By("creating the HcloudNetwork resource with invalid IP range")
			resource := &hcloudv1alpha1.HcloudNetwork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: hcloudv1alpha1.HcloudNetworkSpec{
					Name:    "test-network-invalid",
					IpRange: "invalid-cidr",
					Labels:  map[string]string{},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			mockMgr := &hcloud.MockClient{}
			manager := hcloud.NewNetworkManager(mockMgr)

			By("reconciling the resource")
			reconciler := &HcloudNetworkReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				HCloudMgr: manager,
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())

			By("verifying the Available condition indicates failure")
			updatedResource := &hcloudv1alpha1.HcloudNetwork{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updatedResource)).To(Succeed())
			condition := meta.FindStatusCondition(updatedResource.Status.Conditions, "Available")
			Expect(condition).NotTo(BeNil())
			Expect(condition.Status).To(Equal(metav1.ConditionFalse))
			Expect(condition.Reason).To(Equal("InvalidIPRange"))

			By("cleaning up the resource")
			Expect(k8sClient.Delete(ctx, updatedResource)).To(Succeed())
		})

		It("should handle Hetzner Cloud API errors gracefully", func() {
			const resourceName = "test-api-error"
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: namespace,
			}

			By("creating the HcloudNetwork resource")
			resource := &hcloudv1alpha1.HcloudNetwork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: hcloudv1alpha1.HcloudNetworkSpec{
					Name:    "test-network-api-error",
					IpRange: "10.0.0.0/8",
					Labels:  map[string]string{},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			mockMgr := &hcloud.MockClient{}
			mockMgr.CreateNetworkFunc = func(ctx context.Context, opts hcloudgo.NetworkCreateOpts) (*hcloudgo.Network, error) {
				return nil, fmt.Errorf("API error: rate limit exceeded")
			}

			manager := hcloud.NewNetworkManager(mockMgr)

			By("reconciling the resource")
			reconciler := &HcloudNetworkReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				HCloudMgr: manager,
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())

			By("verifying the Available condition indicates creation failure")
			updatedResource := &hcloudv1alpha1.HcloudNetwork{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updatedResource)).To(Succeed())
			condition := meta.FindStatusCondition(updatedResource.Status.Conditions, "Available")
			Expect(condition).NotTo(BeNil())
			Expect(condition.Status).To(Equal(metav1.ConditionFalse))
			Expect(condition.Reason).To(Equal("CreateFailed"))

			By("cleaning up the resource")
			Expect(k8sClient.Delete(ctx, updatedResource)).To(Succeed())
		})
	})

	Context("Update existing HcloudNetwork", func() {
		const namespace = "default"

		ctx := context.Background()

		It("should successfully update network name and labels", func() {
			const resourceName = "test-update-network"
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: namespace,
			}

			By("creating the HcloudNetwork resource")
			resource := &hcloudv1alpha1.HcloudNetwork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: hcloudv1alpha1.HcloudNetworkSpec{
					Name:    "test-network-v1",
					IpRange: "10.0.0.0/8",
					Labels: map[string]string{
						"version": "v1",
					},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			existingNetwork := &hcloudgo.Network{
				ID:   54321,
				Name: "test-network-v1",
				IPRange: &net.IPNet{
					IP:   net.IPv4(10, 0, 0, 0),
					Mask: net.IPv4Mask(255, 0, 0, 0),
				},
				Labels: map[string]string{"version": "v1"},
			}

			mockMgr := &hcloud.MockClient{}
			mockMgr.GetNetworkByIdFunc = func(ctx context.Context, id int64) (*hcloudgo.Network, error) {
				return existingNetwork, nil
			}
			mockMgr.UpdateNetworkFunc = func(ctx context.Context, id int64, opts hcloudgo.NetworkUpdateOpts) (*hcloudgo.Network, error) {
				existingNetwork.Name = opts.Name
				existingNetwork.Labels = opts.Labels
				return existingNetwork, nil
			}

			manager := hcloud.NewNetworkManager(mockMgr)

			By("setting network ID and finalizer")
			resource.Status.NetworkId = 54321
			Expect(k8sClient.Status().Update(ctx, resource)).To(Succeed())

			getResource := &hcloudv1alpha1.HcloudNetwork{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, getResource)).To(Succeed())
			getResource.ObjectMeta.Finalizers = []string{finalizerName}
			Expect(k8sClient.Update(ctx, getResource)).To(Succeed())

			By("updating the resource spec")
			updatedResource := &hcloudv1alpha1.HcloudNetwork{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updatedResource)).To(Succeed())
			updatedResource.Spec.Name = "test-network-v2"
			updatedResource.Spec.Labels = map[string]string{"version": "v2"}
			Expect(k8sClient.Update(ctx, updatedResource)).To(Succeed())

			By("reconciling the resource")
			reconciler := &HcloudNetworkReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				HCloudMgr: manager,
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("verifying the Available condition is still true")
			finalResource := &hcloudv1alpha1.HcloudNetwork{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, finalResource)).To(Succeed())
			condition := meta.FindStatusCondition(finalResource.Status.Conditions, "Available")
			Expect(condition).NotTo(BeNil())
			Expect(condition.Status).To(Equal(metav1.ConditionTrue))
			Expect(condition.Reason).To(Equal("NetworkReady"))

			By("cleaning up the resource")
			Expect(k8sClient.Delete(ctx, finalResource)).To(Succeed())
		})
	})

	Context("Delete HcloudNetwork with finalizer", func() {
		const namespace = "default"

		ctx := context.Background()

		It("should delete network from Hetzner Cloud and remove finalizer", func() {
			const resourceName = "test-delete-success"
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: namespace,
			}

			By("creating the HcloudNetwork resource with network ID")
			resource := &hcloudv1alpha1.HcloudNetwork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: hcloudv1alpha1.HcloudNetworkSpec{
					Name:    "test-network-delete",
					IpRange: "10.0.0.0/8",
					Labels:  map[string]string{},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			mockMgr := &hcloud.MockClient{}
			mockMgr.DeleteNetworkFunc = func(ctx context.Context, id int64) error {
				return nil
			}

			manager := hcloud.NewNetworkManager(mockMgr)

			By("setting network ID and finalizer")
			resource.Status.NetworkId = 99999
			Expect(k8sClient.Status().Update(ctx, resource)).To(Succeed())

			getResource := &hcloudv1alpha1.HcloudNetwork{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, getResource)).To(Succeed())
			getResource.ObjectMeta.Finalizers = []string{finalizerName}
			Expect(k8sClient.Update(ctx, getResource)).To(Succeed())

			By("initiating deletion of the resource")
			Expect(k8sClient.Delete(ctx, getResource)).To(Succeed())

			By("reconciling the resource")
			reconciler := &HcloudNetworkReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				HCloudMgr: manager,
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("verifying finalizer was removed and resource is gone")
			deletedResource := &hcloudv1alpha1.HcloudNetwork{}
			err = k8sClient.Get(ctx, typeNamespacedName, deletedResource)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		It("should handle deletion errors gracefully", func() {
			const resourceName = "test-delete-error"
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: namespace,
			}

			By("creating the HcloudNetwork resource")
			resource := &hcloudv1alpha1.HcloudNetwork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: hcloudv1alpha1.HcloudNetworkSpec{
					Name:    "test-network-del-error",
					IpRange: "10.0.0.0/8",
					Labels:  map[string]string{},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			mockMgr := &hcloud.MockClient{}
			mockMgr.DeleteNetworkFunc = func(ctx context.Context, id int64) error {
				return fmt.Errorf("API error: network in use")
			}

			manager := hcloud.NewNetworkManager(mockMgr)

			By("setting network ID and finalizer")
			resource.Status.NetworkId = 99999
			Expect(k8sClient.Status().Update(ctx, resource)).To(Succeed())

			getResource := &hcloudv1alpha1.HcloudNetwork{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, getResource)).To(Succeed())
			getResource.ObjectMeta.Finalizers = []string{finalizerName}
			Expect(k8sClient.Update(ctx, getResource)).To(Succeed())

			By("initiating deletion of the resource")
			Expect(k8sClient.Delete(ctx, getResource)).To(Succeed())

			By("reconciling the resource")
			reconciler := &HcloudNetworkReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				HCloudMgr: manager,
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())

			By("verifying the DeletionFailed condition was set")
			failedResource := &hcloudv1alpha1.HcloudNetwork{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, failedResource)).To(Succeed())
			condition := meta.FindStatusCondition(failedResource.Status.Conditions, "Available")
			Expect(condition).NotTo(BeNil())
			Expect(condition.Status).To(Equal(metav1.ConditionFalse))
			Expect(condition.Reason).To(Equal("DeletionFailed"))

			By("verifying finalizer still exists (not removed on error)")
			Expect(failedResource.ObjectMeta.Finalizers).To(ContainElement(finalizerName))

			By("cleaning up the resource")
			Expect(k8sClient.Delete(ctx, failedResource)).To(Succeed())
		})
	})

	Context("Verify network exists", func() {
		const namespace = "default"

		ctx := context.Background()

		It("should handle network verification failure gracefully", func() {
			const resourceName = "test-verify-network"
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: namespace,
			}

			By("creating the HcloudNetwork resource with network ID")
			resource := &hcloudv1alpha1.HcloudNetwork{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: hcloudv1alpha1.HcloudNetworkSpec{
					Name:    "test-network-verify",
					IpRange: "10.0.0.0/8",
					Labels:  map[string]string{},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			mockMgr := &hcloud.MockClient{}
			mockMgr.GetNetworkByIdFunc = func(ctx context.Context, id int64) (*hcloudgo.Network, error) {
				return nil, fmt.Errorf("API error: network not found")
			}

			manager := hcloud.NewNetworkManager(mockMgr)

			By("setting network ID")
			resource.Status.NetworkId = 12345
			Expect(k8sClient.Status().Update(ctx, resource)).To(Succeed())

			By("reconciling the resource")
			reconciler := &HcloudNetworkReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				HCloudMgr: manager,
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())

			By("verifying the VerifyFailed condition was set")
			updatedResource := &hcloudv1alpha1.HcloudNetwork{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updatedResource)).To(Succeed())
			condition := meta.FindStatusCondition(updatedResource.Status.Conditions, "Available")
			Expect(condition).NotTo(BeNil())
			Expect(condition.Status).To(Equal(metav1.ConditionFalse))
			Expect(condition.Reason).To(Equal("VerifyFailed"))

			By("cleaning up the resource")
			Expect(k8sClient.Delete(ctx, updatedResource)).To(Succeed())
		})
	})
})
