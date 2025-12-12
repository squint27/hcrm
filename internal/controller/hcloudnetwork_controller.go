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

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	hcloudv1alpha1 "bunskin.com/hcrm/api/v1alpha1"
	"bunskin.com/hcrm/pkg/hcloud"
)

const (
	// finalizerName is the name of the finalizer for HcloudNetwork resources
	finalizerName = "hcloud.bunskin.com/finalizer"
)

// HcloudNetworkReconciler reconciles a HcloudNetwork object
type HcloudNetworkReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	NetworkClient hcloud.NetworkClient
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
	var hcloudNetwork hcloudv1alpha1.HcloudNetwork
	if err := r.Get(ctx, req.NamespacedName, &hcloudNetwork); err != nil {
		// object does not exist, nothing to do
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.V(1).Info("Reconciling HcloudNetwork", "name", hcloudNetwork.Name, "namespace", hcloudNetwork.Namespace)

	// Handle deletion with finalizer
	if hcloudNetwork.DeletionTimestamp != nil {
		log.V(1).Info("HcloudNetwork resource is being deleted", "name", hcloudNetwork.Name)

		// Check if finalizer exists
		if controllerutil.ContainsFinalizer(&hcloudNetwork, finalizerName) {
			// Delete the network from Hetzner Cloud if it exists
			if hcloudNetwork.Status.NetworkId != 0 {
				log.V(1).Info("Fetching Hetzner Cloud network for deletion", "networkId", hcloudNetwork.Status.NetworkId)
				network, response, err := r.NetworkClient.GetNetworkById(ctx, int64(hcloudNetwork.Status.NetworkId))
				if err != nil {
					log.Error(err, "Failed to get network from Hetzner Cloud", "networkId", hcloudNetwork.Status.NetworkId)
					meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
						Type:               "Available",
						Status:             metav1.ConditionFalse,
						ObservedGeneration: hcloudNetwork.Generation,
						Reason:             "DeletionFailed",
						Message:            fmt.Sprintf("Failed to get network for deletion: %v. %v", err, response),
					})
					if updateErr := r.Status().Update(ctx, &hcloudNetwork); updateErr != nil {
						log.Error(updateErr, "Failed to update HcloudNetwork status")
					}

					return ctrl.Result{}, err
				}

				if network != nil {
					// Delete the network
					log.V(1).Info("Deleting Hetzner Cloud network", "networkId", hcloudNetwork.Status.NetworkId)
					response, err := r.NetworkClient.DeleteNetwork(ctx, network)
					if err != nil {
						log.Error(err, "Failed to delete network from Hetzner Cloud", "networkId", hcloudNetwork.Status.NetworkId)
						meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
							Type:               "Available",
							Status:             metav1.ConditionFalse,
							ObservedGeneration: hcloudNetwork.Generation,
							Reason:             "DeletionFailed",
							Message:            fmt.Sprintf("Failed to delete network from Hetzner Cloud: %v. %v", err, response),
						})
						if updateErr := r.Status().Update(ctx, &hcloudNetwork); updateErr != nil {
							log.Error(updateErr, "Failed to update HcloudNetwork status")
						}
						return ctrl.Result{}, err
					}

					log.V(1).Info("Successfully deleted Hetzner Cloud network", "networkId", hcloudNetwork.Status.NetworkId)
				} else {
					log.V(1).Info("Network not found in Hetzner Cloud, nothing to delete", "networkId", hcloudNetwork.Status.NetworkId)
				}
			}

			// Remove finalizer
			controllerutil.RemoveFinalizer(&hcloudNetwork, finalizerName)
			if err := r.Update(ctx, &hcloudNetwork); err != nil {
				log.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}
			log.V(1).Info("Finalizer removed, resource deletion complete", "name", hcloudNetwork.Name)
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(&hcloudNetwork, finalizerName) {
		controllerutil.AddFinalizer(&hcloudNetwork, finalizerName)
		if err := r.Update(ctx, &hcloudNetwork); err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		log.V(1).Info("Finalizer added", "name", hcloudNetwork.Name)
	}

	// Adopt existing network if it exists
	log.V(1).Info("Checking for existing network in Hetzner Cloud by name", "name", hcloudNetwork.Spec.Name)
	network, response, err := r.NetworkClient.GetNetworkByName(ctx, hcloudNetwork.Spec.Name)
	if err != nil {
		log.Error(err, "Failed to get network from Hetzner Cloud by name", "name", hcloudNetwork.Spec.Name)
		meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             metav1.ConditionFalse,
			ObservedGeneration: hcloudNetwork.Generation,
			Reason:             "GetNetworkFailed",
			Message:            fmt.Sprintf("Failed to get network from Hetzner Cloud by name: %v. %v", err, response),
		})
		if updateErr := r.Status().Update(ctx, &hcloudNetwork); updateErr != nil {
			log.Error(updateErr, "Failed to update HcloudNetwork status")
		}
		return ctrl.Result{}, err
	}

	if network != nil {
		log.V(1).Info("Found existing network in Hetzner Cloud", "networkId", network.ID)

		// Update the resource status with the network ID and conditions
		hcloudNetwork.Status.NetworkId = int(network.ID)
		meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: hcloudNetwork.Generation,
			Reason:             "NetworkReady",
			Message:            fmt.Sprintf("Network found in Hetzner Cloud with ID: %d", network.ID),
		})
		hcloudNetwork.Status.ObservedGeneration = hcloudNetwork.Generation
		if err := r.Status().Update(ctx, &hcloudNetwork); err != nil {
			log.Error(err, "Failed to update HcloudNetwork status")
			return ctrl.Result{}, err
		}
	}

	log.V(1).Info("HcloudNetwork resource reconciled successfully", "name", hcloudNetwork.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HcloudNetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hcloudv1alpha1.HcloudNetwork{}).
		Named("hcloudnetwork").
		Complete(r)
}
