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

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	hcloudv1alpha1 "bunskin.com/hcrm/api/v1alpha1"
	"bunskin.com/hcrm/pkg/hcloud"
)

const (
	// finalizerName is the name of the finalizer for HcloudNetwork resources
	finalizerName = "hcloud.bunskin.com/finalizer"
	// syncPolicy is the annotation key for the sync policy
	syncPolicy = "hcloud.bunskin.com/sync-policy"
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

func (r *HcloudNetworkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.Log.WithName("hcloudnetwork-controller")

	// Fetch the HcloudNetwork resource
	var hcloudNetwork hcloudv1alpha1.HcloudNetwork
	if err := r.Get(ctx, req.NamespacedName, &hcloudNetwork); err != nil {
		// object does not exist, nothing to do
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	originalStatus := hcloudNetwork.Status.DeepCopy()

	defer func() {
		// Update status if it has changed
		if !equality.Semantic.DeepEqual(originalStatus, &hcloudNetwork.Status) {
			if err := r.Status().Update(ctx, &hcloudNetwork); err != nil {
				log.Error(err, "Failed to update HcloudNetwork status", "name", hcloudNetwork.Name)
			}
		}
	}()

	log.Info("Reconciling HcloudNetwork", "name", hcloudNetwork.Name, "namespace", hcloudNetwork.Namespace)

	// Handle deletion with finalizer
	if hcloudNetwork.DeletionTimestamp != nil {
		log.Info("HcloudNetwork resource is being deleted", "name", hcloudNetwork.Name)

		// Check if finalizer exists
		if controllerutil.ContainsFinalizer(&hcloudNetwork, finalizerName) {
			// Delete the network from Hetzner Cloud if it exists
			if hcloudNetwork.Status.NetworkId != 0 {
				log.Info("Fetching Hetzner Cloud network for deletion", "networkId", hcloudNetwork.Status.NetworkId)
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

					return ctrl.Result{}, err
				}

				if network != nil {
					// Delete the network
					log.Info("Deleting Hetzner Cloud network", "networkId", hcloudNetwork.Status.NetworkId)
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

						return ctrl.Result{}, err
					}

					log.Info("Successfully deleted Hetzner Cloud network", "networkId", hcloudNetwork.Status.NetworkId)
				} else {
					log.Info("Network not found in Hetzner Cloud, nothing to delete", "networkId", hcloudNetwork.Status.NetworkId)
				}
			}

			// Remove finalizer
			controllerutil.RemoveFinalizer(&hcloudNetwork, finalizerName)
			if err := r.Update(ctx, &hcloudNetwork); err != nil {
				log.Error(err, "Failed to remove finalizer", "name", hcloudNetwork.Name)
				return ctrl.Result{}, err
			}
			log.Info("Finalizer removed, resource deletion complete", "name", hcloudNetwork.Name)
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(&hcloudNetwork, finalizerName) {
		controllerutil.AddFinalizer(&hcloudNetwork, finalizerName)
		if err := r.Update(ctx, &hcloudNetwork); err != nil {
			log.Error(err, "Failed to add finalizer", "name", hcloudNetwork.Name)
			return ctrl.Result{}, err
		}

		log.Info("Finalizer added", "name", hcloudNetwork.Name)
	}

	// Initialize annotations if not present
	if hcloudNetwork.Annotations == nil {
		hcloudNetwork.Annotations = make(map[string]string)
	}

	// Add sync policy annotation if not present
	if hcloudNetwork.Annotations[syncPolicy] == "" {
		log.Info("Adding sync policy annotation", "name", hcloudNetwork.Name)
		hcloudNetwork.Annotations[syncPolicy] = "manage"
		if err := r.Update(ctx, &hcloudNetwork); err != nil {
			log.Error(err, "Failed to add sync policy annotation", "name", hcloudNetwork.Name)
			return ctrl.Result{}, err
		}

		log.Info("Sync policy annotation added", "name", hcloudNetwork.Name)

	}

	// Adopt existing network if it exists
	log.Info("Checking for existing network in Hetzner Cloud by name", "name", hcloudNetwork.Spec.Name)
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

		return ctrl.Result{}, err
	}

	if network != nil {
		log.Info("Found existing network in Hetzner Cloud", "networkId", network.ID)

		// Evaluate if the network spec matches the existing network
		needsLabelsUpdate := false
		needsCidrUpdate := false
		if hcloudNetwork.Spec.IpRange != network.IPRange.String() {
			log.Info("Network IP range differs, updating", "current", network.IPRange, "desired", hcloudNetwork.Spec.IpRange)
			needsCidrUpdate = true
		} else if hcloudNetwork.Spec.Labels != nil && !equality.Semantic.DeepEqual(hcloudNetwork.Spec.Labels, network.Labels) {
			log.Info("Network labels differ, updating", "current", network.Labels, "desired", hcloudNetwork.Spec.Labels)
			needsLabelsUpdate = true
		}

		if needsLabelsUpdate {
			_, response, err := r.NetworkClient.UpdateNetworkLabels(ctx, network, hcloudNetwork.Spec.Labels)
			if err != nil {
				log.Error(err, "Failed to update network labels in Hetzner Cloud", "networkId", network.ID)
				meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
					Type:               "Available",
					Status:             metav1.ConditionFalse,
					ObservedGeneration: hcloudNetwork.Generation,
					Reason:             "UpdateNetworkFailed",
					Message:            fmt.Sprintf("Failed to update network in Hetzner Cloud: %v. %v", err, response),
				})

				return ctrl.Result{}, err
			}
		} else if needsCidrUpdate {
			_, response, err = r.NetworkClient.UpdateNetworkCidr(ctx, network, hcloudNetwork.Spec.IpRange)
			if err != nil {
				log.Error(err, "Failed to update network CIDR in Hetzner Cloud", "networkId", network.ID)
				meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
					Type:               "Available",
					Status:             metav1.ConditionFalse,
					ObservedGeneration: hcloudNetwork.Generation,
					Reason:             "UpdateNetworkFailed",
					Message:            fmt.Sprintf("Failed to update network in Hetzner Cloud: %v. %v", err, response),
				})

				return ctrl.Result{}, err
			}

			log.Info("Successfully updated network in Hetzner Cloud", "networkId", network.ID)
		} else {
			log.Info("No updates required for existing network", "networkId", network.ID)
		}

		// Update the resource status with the network ID and conditions
		hcloudNetwork.Status.NetworkId = int(network.ID)
		meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: hcloudNetwork.Generation,
			Reason:             "NetworkReady",
			Message:            fmt.Sprintf("Network ID %d updated successfully", network.ID),
		})
		hcloudNetwork.Status.ObservedGeneration = hcloudNetwork.Generation

	} else {
		log.Info("Network not found in Hetzner Cloud, creating new network", "name", hcloudNetwork.Spec.Name)

		network, response, err := r.NetworkClient.CreateNetwork(ctx, hcloudNetwork.Spec.Name, hcloudNetwork.Spec.IpRange, hcloudNetwork.Spec.Labels)
		if err != nil {
			log.Error(err, "Failed to create network in Hetzner Cloud", "name", hcloudNetwork.Spec.Name)
			meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
				Type:               "Available",
				Status:             metav1.ConditionFalse,
				ObservedGeneration: hcloudNetwork.Generation,
				Reason:             "CreateNetworkFailed",
				Message:            fmt.Sprintf("Failed to create network in Hetzner Cloud: %v. %v", err, response),
			})

			return ctrl.Result{}, err
		}

		log.Info("Successfully created network in Hetzner Cloud", "networkId", network.ID)

		hcloudNetwork.Status.NetworkId = int(network.ID)
		meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: hcloudNetwork.Generation,
			Reason:             "NetworkCreated",
			Message:            fmt.Sprintf("Network created in Hetzner Cloud with ID: %d", network.ID),
		})
		hcloudNetwork.Status.ObservedGeneration = hcloudNetwork.Generation
	}

	log.Info("HcloudNetwork resource reconciled successfully", "name", hcloudNetwork.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HcloudNetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hcloudv1alpha1.HcloudNetwork{}).
		Named("hcloudnetwork").
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
