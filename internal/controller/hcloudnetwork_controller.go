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
	"k8s.io/client-go/tools/record"
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
	Recorder      record.EventRecorder
}

// +kubebuilder:rbac:groups=hcloud.bunskin.com,resources=hcloudnetworks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hcloud.bunskin.com,resources=hcloudnetworks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hcloud.bunskin.com,resources=hcloudnetworks/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *HcloudNetworkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.Log.WithName("hcloudnetwork-controller")

	// Fetch the HcloudNetwork resource
	var hcloudNetwork hcloudv1alpha1.HcloudNetwork
	if err := r.Get(ctx, req.NamespacedName, &hcloudNetwork); err != nil {
		// object does not exist, nothing to do
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// defer func() {
	// 	// Update status if it has changed
	// 	if !equality.Semantic.DeepEqual(originalStatus, &hcloudNetwork.Status) {
	// 		if err := r.Status().Update(ctx, &hcloudNetwork); err != nil {
	// 			log.Error(err, "Failed to update HcloudNetwork status", "name", hcloudNetwork.Name)
	// 		}
	// 	}
	// }()

	log.Info("Reconciling HcloudNetwork", "name", hcloudNetwork.Name, "namespace", hcloudNetwork.Namespace)
	meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
		Type:               "Available",
		Status:             metav1.ConditionFalse,
		ObservedGeneration: hcloudNetwork.Generation,
		Reason:             "Progressing",
		Message:            "HcloudNetwork resource reconciliation in progress",
	})
	if err := r.Status().Update(ctx, &hcloudNetwork); err != nil {
		log.Error(err, "Failed to update HcloudNetwork status", "name", hcloudNetwork.Name)
		return ctrl.Result{}, err
	}

	// Handle deletion with finalizer
	if hcloudNetwork.DeletionTimestamp != nil {
		log.Info("HcloudNetwork resource is being deleted", "name", hcloudNetwork.Name)
		meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             metav1.ConditionFalse,
			ObservedGeneration: hcloudNetwork.Generation,
			Reason:             "Deleting",
			Message:            "HcloudNetwork resource is being deleted",
		})
		if err := r.Status().Update(ctx, &hcloudNetwork); err != nil {
			log.Error(err, "Failed to update HcloudNetwork status", "name", hcloudNetwork.Name)
			return ctrl.Result{}, err
		}

		// Check if finalizer exists
		if controllerutil.ContainsFinalizer(&hcloudNetwork, finalizerName) {
			// Delete the network from Hetzner Cloud if it exists
			if hcloudNetwork.Status.NetworkId != 0 && hcloudNetwork.Annotations[syncPolicy] != "orphan" {
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
					if err := r.Status().Update(ctx, &hcloudNetwork); err != nil {
						log.Error(err, "Failed to update HcloudNetwork status", "name", hcloudNetwork.Name)
					}

					r.Recorder.Eventf(&hcloudNetwork, "Warning", "DeletionFailed", "Failed to get network %d for deletion", hcloudNetwork.Status.NetworkId)

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
						if err := r.Status().Update(ctx, &hcloudNetwork); err != nil {
							log.Error(err, "Failed to update HcloudNetwork status", "name", hcloudNetwork.Name)
						}

						r.Recorder.Eventf(&hcloudNetwork, "Warning", "Failed to delete network %s from Hetzner cloud", hcloudNetwork.Spec.Name)

						return ctrl.Result{}, err
					}

					log.Info("Successfully deleted Hetzner Cloud network", "networkId", hcloudNetwork.Status.NetworkId)
					r.Recorder.Eventf(&hcloudNetwork, "Normal", "Deleted", "HcloudNetwork %s deleted successfully", hcloudNetwork.Spec.Name)
				} else {
					log.Info("Network not found in Hetzner Cloud, nothing to delete", "networkId", hcloudNetwork.Status.NetworkId)
				}
			} else if hcloudNetwork.Annotations[syncPolicy] == "orphan" {
				log.Info("Sync policy is set to orphan, will not remove cloud resource")
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
	}

	// Add finalizer if not present and sync policy supports it
	if !controllerutil.ContainsFinalizer(&hcloudNetwork, finalizerName) && hcloudNetwork.Annotations[syncPolicy] != "read-only" {
		log.Info("Adding finalizer", "name", hcloudNetwork.Name)
		controllerutil.AddFinalizer(&hcloudNetwork, finalizerName)
		if err := r.Update(ctx, &hcloudNetwork); err != nil {
			log.Error(err, "Failed to add finalizer", "name", hcloudNetwork.Name)
			return ctrl.Result{}, err
		}
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
			Reason:             "Failed",
			Message:            fmt.Sprintf("Failed to get network from Hetzner Cloud by name: %v. %v", err, response),
		})
		if err := r.Status().Update(ctx, &hcloudNetwork); err != nil {
			log.Error(err, "Failed to update HcloudNetwork status", "name", hcloudNetwork.Name)
		}
		r.Recorder.Eventf(&hcloudNetwork, "Warning", "UpdateFailed", "Failed to get network %s from Hetzner cloud", hcloudNetwork.Spec.Name)

		return ctrl.Result{}, err
	}

	if network != nil {
		log.Info("Found existing network in Hetzner Cloud", "networkId", network.ID)

		// Update the existing network if sync policy allows it
		if hcloudNetwork.Annotations[syncPolicy] != "read-only" {
			// Evaluate if the network spec matches the existing network
			needsLabelsUpdate := false
			needsCidrUpdate := false
			if hcloudNetwork.Spec.IpRange != network.IPRange.String() {
				log.Info("Network IP range differs, updating", "current", network.IPRange, "desired", hcloudNetwork.Spec.IpRange)
				needsCidrUpdate = true
			}
			if hcloudNetwork.Spec.Labels != nil && !equality.Semantic.DeepEqual(hcloudNetwork.Spec.Labels, network.Labels) {
				log.Info("Network labels differ, updating", "current", network.Labels, "desired", hcloudNetwork.Spec.Labels)
				needsLabelsUpdate = true
			}

			if needsLabelsUpdate {
				updatedNetwork, response, err := r.NetworkClient.UpdateNetworkLabels(ctx, network, hcloudNetwork.Spec.Labels)
				if err != nil {
					log.Error(err, "Failed to update network labels in Hetzner Cloud", "networkId", network.ID)
					meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
						Type:               "Available",
						Status:             metav1.ConditionFalse,
						ObservedGeneration: hcloudNetwork.Generation,
						Reason:             "Failed",
						Message:            fmt.Sprintf("Failed to update network in Hetzner Cloud: %v. %v", err, response),
					})
					if err := r.Status().Update(ctx, &hcloudNetwork); err != nil {
						log.Error(err, "Failed to update HcloudNetwork status", "name", hcloudNetwork.Name)
					}
					r.Recorder.Eventf(&hcloudNetwork, "Warning", "UpdateFailed", "Failed to update network %s in Hetzner cloud", hcloudNetwork.Spec.Name)

					return ctrl.Result{}, err
				}
				network = updatedNetwork
			}
			if needsCidrUpdate {
				updatedNetwork, response, err := r.NetworkClient.UpdateNetworkCidr(ctx, network, hcloudNetwork.Spec.IpRange)
				if err != nil {
					log.Error(err, "Failed to update network CIDR in Hetzner Cloud", "networkId", network.ID)
					meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
						Type:               "Available",
						Status:             metav1.ConditionFalse,
						ObservedGeneration: hcloudNetwork.Generation,
						Reason:             "Failed",
						Message:            fmt.Sprintf("Failed to update network in Hetzner Cloud: %v. %v", err, response),
					})
					if err := r.Status().Update(ctx, &hcloudNetwork); err != nil {
						log.Error(err, "Failed to update HcloudNetwork status", "name", hcloudNetwork.Name)
					}
					r.Recorder.Eventf(&hcloudNetwork, "Warning", "UpdateFailed", "Failed to update network %s in Hetzner cloud", hcloudNetwork.Spec.Name)

					return ctrl.Result{}, err
				}

				network = updatedNetwork
				log.Info("Successfully updated network in Hetzner Cloud", "networkId", network.ID)
			}
			if !needsLabelsUpdate && !needsCidrUpdate {
				log.Info("No updates required for existing network", "networkId", network.ID)
			}
		} else {
			log.Info("Sync policy is read-only; skipping updates to existing network", "networkId", network.ID)
		}

		// Update the resource status with the network details and conditions
		hcloudNetwork.Status.NetworkId = int(network.ID)
		hcloudNetwork.Status.IpRange = network.IPRange.String()
		hcloudNetwork.Status.Labels = network.Labels
		meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: hcloudNetwork.Generation,
			Reason:             "Ready",
			Message:            fmt.Sprintf("Network ID %d reconciled successfully", network.ID),
		})
		hcloudNetwork.Status.ObservedGeneration = hcloudNetwork.Generation

		if err := r.Status().Update(ctx, &hcloudNetwork); err != nil {
			log.Error(err, "Failed to update HcloudNetwork status", "name", hcloudNetwork.Name)
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(&hcloudNetwork, "Normal", "Ready", "HcloudNetwork updated %d", hcloudNetwork.Status.NetworkId)

	} else if hcloudNetwork.Annotations[syncPolicy] != "read-only" {
		log.Info("Network not found in Hetzner Cloud, creating new network", "name", hcloudNetwork.Spec.Name)

		network, response, err := r.NetworkClient.CreateNetwork(ctx, hcloudNetwork.Spec.Name, hcloudNetwork.Spec.IpRange, hcloudNetwork.Spec.Labels)
		if err != nil {
			log.Error(err, "Failed to create network in Hetzner Cloud", "name", hcloudNetwork.Spec.Name)
			meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
				Type:               "Available",
				Status:             metav1.ConditionFalse,
				ObservedGeneration: hcloudNetwork.Generation,
				Reason:             "Failed",
				Message:            fmt.Sprintf("Failed to create network in Hetzner Cloud: %v. %v", err, response),
			})
			if err := r.Status().Update(ctx, &hcloudNetwork); err != nil {
				log.Error(err, "Failed to update HcloudNetwork status", "name", hcloudNetwork.Name)
			}
			r.Recorder.Eventf(&hcloudNetwork, "Warning", "CreateFailed", "Failed to create network %s in Hetzner cloud", hcloudNetwork.Spec.Name)

			return ctrl.Result{}, err
		}

		log.Info("Successfully created network in Hetzner Cloud", "networkId", network.ID)

		hcloudNetwork.Status.NetworkId = int(network.ID)
		hcloudNetwork.Status.IpRange = network.IPRange.String()
		hcloudNetwork.Status.Labels = network.Labels
		meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: hcloudNetwork.Generation,
			Reason:             "Ready",
			Message:            fmt.Sprintf("Network created in Hetzner Cloud with ID: %d", network.ID),
		})
		hcloudNetwork.Status.ObservedGeneration = hcloudNetwork.Generation

		if err := r.Status().Update(ctx, &hcloudNetwork); err != nil {
			log.Error(err, "Failed to update HcloudNetwork status", "name", hcloudNetwork.Name)
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(&hcloudNetwork, "Normal", "Ready", "HcloudNetwork created %d", hcloudNetwork.Status.NetworkId)
	} else {
		log.Info("Network not found in Hetzner Cloud and sync policy is read-only; skipping creation", "name", hcloudNetwork.Spec.Name)
		meta.SetStatusCondition(&hcloudNetwork.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             metav1.ConditionFalse,
			ObservedGeneration: hcloudNetwork.Generation,
			Reason:             "Failed",
			Message:            "Network not found in Hetzner Cloud and sync policy is read-only",
		})
		if err := r.Status().Update(ctx, &hcloudNetwork); err != nil {
			log.Error(err, "Failed to update HcloudNetwork status", "name", hcloudNetwork.Name)
			return ctrl.Result{}, err
		}
		r.Recorder.Eventf(&hcloudNetwork, "Warning", "Failed", "Network %s not found in Hetzner cloud", hcloudNetwork.Spec.Name)
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
