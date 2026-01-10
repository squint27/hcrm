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
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	hcloudv1alpha1 "bunskin.com/hcrm/api/v1alpha1"
	"bunskin.com/hcrm/pkg/hcloud"
)

// HcloudDnsZoneReconciler reconciles a HcloudDnsZone object
type HcloudDnsZoneReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	DnsZoneClient hcloud.DnsZoneClient
	Recorder      record.EventRecorder
}

// +kubebuilder:rbac:groups=hcloud.bunskin.com,resources=hclouddnszones,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hcloud.bunskin.com,resources=hclouddnszones/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hcloud.bunskin.com,resources=hclouddnszones/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HcloudDnsZone object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *HcloudDnsZoneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.Log.WithName("hclouddnszone-controller")

	// Fetch the HcloudDnsZone resource
	var hcloudDnsZone hcloudv1alpha1.HcloudDnsZone
	if err := r.Get(ctx, req.NamespacedName, &hcloudDnsZone); err != nil {
		// object does not exist, nothing to do
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Reconciling HcloudDnsZone", "name", hcloudDnsZone.Name, "namespace", hcloudDnsZone.Namespace)
	meta.SetStatusCondition(&hcloudDnsZone.Status.Conditions, metav1.Condition{
		Type:               "Available",
		Status:             metav1.ConditionFalse,
		ObservedGeneration: hcloudDnsZone.Generation,
		Reason:             "Progressing",
		Message:            "HcloudDnsZone resource reconciliation in progress",
	})
	if err := r.Status().Update(ctx, &hcloudDnsZone); err != nil {
		log.Error(err, "Failed to update HcloudDnsZone status", "name", hcloudDnsZone.Name)
		return ctrl.Result{}, err
	}

	// Handle deletion with finalizer
	if hcloudDnsZone.DeletionTimestamp != nil {
		log.Info("HcloudDnsZone resource is being deleted", "name", hcloudDnsZone.Name)
		meta.SetStatusCondition(&hcloudDnsZone.Status.Conditions, metav1.Condition{
			Type:               "Available",
			Status:             metav1.ConditionFalse,
			ObservedGeneration: hcloudDnsZone.Generation,
			Reason:             "Deleting",
			Message:            "HcloudDnsZone resource is being deleted",
		})
		if err := r.Status().Update(ctx, &hcloudDnsZone); err != nil {
			log.Error(err, "Failed to update HcloudDnsZone status", "name", hcloudDnsZone.Name)
			return ctrl.Result{}, err
		}

		// Check if finalizer exists
		if controllerutil.ContainsFinalizer(&hcloudDnsZone, finalizerName) {
			// Delete the DNS zone from Hetzner Cloud if it exists
			if hcloudDnsZone.Status.ZoneId != 0 && hcloudDnsZone.Annotations[syncPolicy] != "orphan" {
				log.Info("Fetching Hetzner Cloud DNS zone for deletion", "dnsZoneId", hcloudDnsZone.Status.ZoneId)
				zone, response, err := r.DnsZoneClient.GetZoneById(ctx, int64(hcloudDnsZone.Status.ZoneId))
				if err != nil {
					log.Error(err, "Failed to get dns zone from Hetzner Cloud", "zoneId", hcloudDnsZone.Status.ZoneId)
					meta.SetStatusCondition(&hcloudDnsZone.Status.Conditions, metav1.Condition{
						Type:               "Available",
						Status:             metav1.ConditionFalse,
						ObservedGeneration: hcloudDnsZone.Generation,
						Reason:             "DeletionFailed",
						Message:            fmt.Sprintf("Failed to get network for deletion: %v. %v", err, response),
					})
					if err := r.Status().Update(ctx, &hcloudDnsZone); err != nil {
						log.Error(err, "Failed to update HcloudDnsZone status", "name", hcloudDnsZone.Name)
					}

					r.Recorder.Eventf(&hcloudDnsZone, "Warning", "DeletionFailed", "Failed to get dns zone %d for deletion", hcloudDnsZone.Status.ZoneId)

					return ctrl.Result{}, err
				}

				if zone != nil {
					// Delete the dns zone
					log.Info("Deleting Hetzner Cloud dns zone", "zoneId", hcloudDnsZone.Status.ZoneId)
					response, err := r.DnsZoneClient.DeleteZone(ctx, zone)
					if err != nil {
						log.Error(err, "Failed to delete dns zone from Hetzner Cloud", "zoneId", hcloudDnsZone.Status.ZoneId)
						meta.SetStatusCondition(&hcloudDnsZone.Status.Conditions, metav1.Condition{
							Type:               "Available",
							Status:             metav1.ConditionFalse,
							ObservedGeneration: hcloudDnsZone.Generation,
							Reason:             "DeletionFailed",
							Message:            fmt.Sprintf("Failed to delete dns zone from Hetzner Cloud: %v. %v", err, response),
						})
						if err := r.Status().Update(ctx, &hcloudDnsZone); err != nil {
							log.Error(err, "Failed to update HcloudDnsZone status", "name", hcloudDnsZone.Name)
						}

						r.Recorder.Eventf(&hcloudDnsZone, "Warning", "Failed to delete dns zone %s from Hetzner cloud", hcloudDnsZone.Spec.Name)

						return ctrl.Result{}, err
					}

					log.Info("Successfully deleted Hetzner Cloud dns zone", "zoneId", hcloudDnsZone.Status.ZoneId)
					r.Recorder.Eventf(&hcloudDnsZone, "Normal", "Deleted", "HcloudDnsZone %s deleted successfully", hcloudDnsZone.Spec.Name)
				} else {
					log.Info("DNS zone not found in Hetzner Cloud, nothing to delete", "zoneId", hcloudDnsZone.Status.ZoneId)
				}
			} else if hcloudDnsZone.Annotations[syncPolicy] == "orphan" {
				log.Info("Sync policy is set to orphan, will not remove cloud resource")
			}

			// Remove finalizer
			controllerutil.RemoveFinalizer(&hcloudDnsZone, finalizerName)
			if err := r.Update(ctx, &hcloudDnsZone); err != nil {
				log.Error(err, "Failed to remove finalizer", "name", hcloudDnsZone.Name)
				return ctrl.Result{}, err
			}
			log.Info("Finalizer removed, resource deletion complete", "name", hcloudDnsZone.Name)
		}
		return ctrl.Result{}, nil
	}

	// Initialize annotations if not present
	if hcloudDnsZone.Annotations == nil {
		hcloudDnsZone.Annotations = make(map[string]string)
	}

	// Add sync policy annotation if not present
	if hcloudDnsZone.Annotations[syncPolicy] == "" {
		log.Info("Adding sync policy annotation", "name", hcloudDnsZone.Name)
		hcloudDnsZone.Annotations[syncPolicy] = "manage"
		if err := r.Update(ctx, &hcloudDnsZone); err != nil {
			log.Error(err, "Failed to add sync policy annotation", "name", hcloudDnsZone.Name)
			return ctrl.Result{}, err
		}
	}

	// Add finalizer if not present and sync policy supports it
	if !controllerutil.ContainsFinalizer(&hcloudDnsZone, finalizerName) && hcloudDnsZone.Annotations[syncPolicy] != "read-only" {
		log.Info("Adding finalizer", "name", hcloudDnsZone.Name)
		controllerutil.AddFinalizer(&hcloudDnsZone, finalizerName)
		if err := r.Update(ctx, &hcloudDnsZone); err != nil {
			log.Error(err, "Failed to add finalizer", "name", hcloudDnsZone.Name)
			return ctrl.Result{}, err
		}
	}

	log.Info("HcloudDnsZone reconciled successfully", "name", hcloudDnsZone.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HcloudDnsZoneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hcloudv1alpha1.HcloudDnsZone{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Named("hclouddnszone").
		Complete(r)
}
