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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	hcloudv1alpha1 "bunskin.com/hcrm/api/v1alpha1"
	"bunskin.com/hcrm/pkg/hcloud"
	hcloudgo "github.com/hetznercloud/hcloud-go/v2/hcloud"
)

const (
	// finalizerName is the name of the finalizer for HcloudNetwork resources
	finalizerName = "hcloud.bunskin.com/finalizer"
)

// HcloudNetworkReconciler reconciles a HcloudNetwork object
type HcloudNetworkReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	HCloudMgr *hcloud.NetworkManager
}

// +kubebuilder:rbac:groups=hcloud.bunskin.com,resources=hcloudnetworks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hcloud.bunskin.com,resources=hcloudnetworks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hcloud.bunskin.com,resources=hcloudnetworks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// It reconciles HcloudNetwork CRD resources by:
// 1. Fetching the HcloudNetwork resource from the cluster
// 2. Creating, updating, or deleting Hetzner Cloud networks as needed
// 3. Updating the resource status with the network ID and conditions
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *HcloudNetworkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the HcloudNetwork resource
	hcloudNetwork := &hcloudv1alpha1.HcloudNetwork{}
	err := r.Get(ctx, req.NamespacedName, hcloudNetwork)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.V(1).Info("HcloudNetwork resource not found, ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get HcloudNetwork resource")
		return ctrl.Result{}, err
	}

	log.V(1).Info("Reconciling HcloudNetwork", "name", hcloudNetwork.Name, "namespace", hcloudNetwork.Namespace)

	// Defensive check: ensure HCloudMgr is configured to avoid nil pointer dereference
	if r.HCloudMgr == nil {
		err := fmt.Errorf("HCloud manager not configured: HCLOUD_TOKEN missing or manager not injected")
		log.Error(err, "HCloudMgr is nil")
		meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             metav1.ConditionFalse,
			ObservedGeneration: hcloudNetwork.Generation,
			Reason:             "NotConfigured",
			Message:            err.Error(),
		})
		// best-effort status update; ignore error here to keep original error visible
		_ = r.Status().Update(ctx, hcloudNetwork)
		return ctrl.Result{}, err
	}

	// Handle deletion with finalizer
	if hcloudNetwork.ObjectMeta.DeletionTimestamp != nil {
		log.V(1).Info("HcloudNetwork resource is being deleted", "name", hcloudNetwork.Name)

		// Check if finalizer exists
		if controllerutil.ContainsFinalizer(hcloudNetwork, finalizerName) {
			// Delete the network from Hetzner Cloud if it exists
			if hcloudNetwork.Status.NetworkId != 0 {
				log.V(1).Info("Deleting Hetzner Cloud network", "networkId", hcloudNetwork.Status.NetworkId)
				err := r.HCloudMgr.DeleteNetwork(ctx, int64(hcloudNetwork.Status.NetworkId))
				if err != nil {
					log.Error(err, "Failed to delete network from Hetzner Cloud", "networkId", hcloudNetwork.Status.NetworkId)
					meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
						Type:               "Available",
						Status:             metav1.ConditionFalse,
						ObservedGeneration: hcloudNetwork.Generation,
						Reason:             "DeletionFailed",
						Message:            fmt.Sprintf("Failed to delete network from Hetzner Cloud: %v", err),
					})
					if updateErr := r.Status().Update(ctx, hcloudNetwork); updateErr != nil {
						log.Error(updateErr, "Failed to update HcloudNetwork status")
					}
					return ctrl.Result{}, err
				}
				log.V(1).Info("Successfully deleted Hetzner Cloud network", "networkId", hcloudNetwork.Status.NetworkId)
			}

			// Remove finalizer
			controllerutil.RemoveFinalizer(hcloudNetwork, finalizerName)
			if err := r.Update(ctx, hcloudNetwork); err != nil {
				log.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}
			log.V(1).Info("Finalizer removed, resource deletion complete", "name", hcloudNetwork.Name)
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(hcloudNetwork, finalizerName) {
		controllerutil.AddFinalizer(hcloudNetwork, finalizerName)
		if err := r.Update(ctx, hcloudNetwork); err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		log.V(1).Info("Finalizer added", "name", hcloudNetwork.Name)
	}

	// Check if the resource has a network ID (already created)
	if hcloudNetwork.Status.NetworkId == 0 {
		log.V(1).Info("Creating new Hetzner Cloud network", "name", hcloudNetwork.Spec.Name)

		// Parse IP range
		_, ipnet, err := net.ParseCIDR(hcloudNetwork.Spec.IpRange)
		if err != nil {
			log.Error(err, "Failed to parse IP range", "ipRange", hcloudNetwork.Spec.IpRange)
			meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
				Type:               "Available",
				Status:             metav1.ConditionFalse,
				ObservedGeneration: hcloudNetwork.Generation,
				Reason:             "InvalidIPRange",
				Message:            fmt.Sprintf("Invalid IP range: %v", err),
			})
			if updateErr := r.Status().Update(ctx, hcloudNetwork); updateErr != nil {
				log.Error(updateErr, "Failed to update HcloudNetwork status")
			}
			return ctrl.Result{}, err
		}

		// Create network in Hetzner Cloud
		network, err := r.HCloudMgr.CreateNetwork(ctx, hcloudgo.NetworkCreateOpts{
			Name:    hcloudNetwork.Spec.Name,
			IPRange: ipnet,
			Labels:  hcloudNetwork.Spec.Labels,
		})
		if err != nil {
			log.Error(err, "Failed to create network in Hetzner Cloud")
			meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
				Type:               "Available",
				Status:             metav1.ConditionFalse,
				ObservedGeneration: hcloudNetwork.Generation,
				Reason:             "CreateFailed",
				Message:            fmt.Sprintf("Failed to create network: %v", err),
			})
			if updateErr := r.Status().Update(ctx, hcloudNetwork); updateErr != nil {
				log.Error(updateErr, "Failed to update HcloudNetwork status")
			}
			return ctrl.Result{}, err
		}

		// Update status with network ID
		hcloudNetwork.Status.NetworkId = int(network.ID)
		hcloudNetwork.Status.ObservedGeneration = hcloudNetwork.Generation
		meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: hcloudNetwork.Generation,
			Reason:             "NetworkCreated",
			Message:            fmt.Sprintf("Hetzner Cloud network created with ID: %d", network.ID),
		})
		if err := r.Status().Update(ctx, hcloudNetwork); err != nil {
			log.Error(err, "Failed to update HcloudNetwork status after creation")
			return ctrl.Result{}, err
		}
		log.V(1).Info("Successfully created Hetzner Cloud network", "networkId", network.ID)
	} else {
		// Network already exists, verify and update if needed
		log.V(1).Info("Network already exists, verifying", "networkId", hcloudNetwork.Status.NetworkId)

		network, err := r.HCloudMgr.GetNetworkById(ctx, int64(hcloudNetwork.Status.NetworkId))
		if err != nil {
			log.Error(err, "Failed to get network from Hetzner Cloud", "networkId", hcloudNetwork.Status.NetworkId)
			meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
				Type:               "Available",
				Status:             metav1.ConditionFalse,
				ObservedGeneration: hcloudNetwork.Generation,
				Reason:             "VerifyFailed",
				Message:            fmt.Sprintf("Failed to verify network: %v", err),
			})
			if updateErr := r.Status().Update(ctx, hcloudNetwork); updateErr != nil {
				log.Error(updateErr, "Failed to update HcloudNetwork status")
			}
			return ctrl.Result{}, err
		}

		// Check if name or labels need to be updated
		needsUpdate := network.Name != hcloudNetwork.Spec.Name ||
			!labelsMatch(network.Labels, hcloudNetwork.Spec.Labels)

		if needsUpdate {
			log.V(1).Info("Updating network in Hetzner Cloud", "networkId", hcloudNetwork.Status.NetworkId)

			_, err := r.HCloudMgr.UpdateNetwork(ctx, int64(hcloudNetwork.Status.NetworkId), hcloudgo.NetworkUpdateOpts{
				Name:   hcloudNetwork.Spec.Name,
				Labels: hcloudNetwork.Spec.Labels,
			})
			if err != nil {
				log.Error(err, "Failed to update network in Hetzner Cloud", "networkId", hcloudNetwork.Status.NetworkId)
				meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
					Type:               "Available",
					Status:             metav1.ConditionFalse,
					ObservedGeneration: hcloudNetwork.Generation,
					Reason:             "UpdateFailed",
					Message:            fmt.Sprintf("Failed to update network: %v", err),
				})
				if updateErr := r.Status().Update(ctx, hcloudNetwork); updateErr != nil {
					log.Error(updateErr, "Failed to update HcloudNetwork status")
				}
				return ctrl.Result{}, err
			}
			log.V(1).Info("Successfully updated network in Hetzner Cloud")
		}

		// Ensure status conditions are set to Available
		meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: hcloudNetwork.Generation,
			Reason:             "NetworkReady",
			Message:            fmt.Sprintf("Hetzner Cloud network with ID %d is ready", network.ID),
		})
		hcloudNetwork.Status.ObservedGeneration = hcloudNetwork.Generation
		if err := r.Status().Update(ctx, hcloudNetwork); err != nil {
			log.Error(err, "Failed to update HcloudNetwork status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// labelsMatch compares two label maps
func labelsMatch(existing map[string]string, desired map[string]string) bool {
	if len(existing) != len(desired) {
		return false
	}
	for key, value := range desired {
		if existing[key] != value {
			return false
		}
	}
	return true
}

// SetupWithManager sets up the controller with the Manager.
func (r *HcloudNetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hcloudv1alpha1.HcloudNetwork{}).
		Named("hcloudnetwork").
		Complete(r)
}
