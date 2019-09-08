// Copyright 2019 The Openstack-Service-Lifecyle Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package osphases

import (
	"context"
	"fmt"

	av1 "github.com/keleustes/oslc-operator/pkg/apis/openstacklcm/v1alpha1"
	deletephasemgr "github.com/keleustes/oslc-operator/pkg/osphases"
	services "github.com/keleustes/oslc-operator/pkg/services"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	crthandler "sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var deletephaselog = logf.Log.WithName("deletephase-controller")

// AddDeletePhaseController creates a new DeletePhase Controller and adds it to
// the Manager. The Manager will set fields on the Controller and Start it when
// the Manager is Started.
func AddDeletePhaseController(mgr manager.Manager) error {
	return addDeletePhase(mgr, newDeletePhaseReconciler(mgr))
}

// newDeletePhaseReconciler returns a new reconcile.Reconciler
func newDeletePhaseReconciler(mgr manager.Manager) reconcile.Reconciler {
	r := &DeletePhaseReconciler{
		PhaseReconciler: PhaseReconciler{
			client:         mgr.GetClient(),
			scheme:         mgr.GetScheme(),
			recorder:       mgr.GetEventRecorderFor("deletephase-recorder"),
			managerFactory: deletephasemgr.NewManagerFactory(mgr),
			// reconcilePeriod: flags.ReconcilePeriod,
		},
	}
	return r
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func addDeletePhase(mgr manager.Manager, r reconcile.Reconciler) error {

	// Create a new controller
	c, err := controller.New("deletephase-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource DeletePhase
	// EnqueueRequestForObject enqueues a Request containing the Name and Namespace of the object
	// that is the source of the Event. (e.g. the created / deleted / updated objects Name and Namespace).
	err = c.Watch(&source.Kind{Type: &av1.DeletePhase{}}, &crthandler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource (described in the yaml file/chart) and requeue the owner DeletePhase
	// EnqueueRequestForOwner enqueues Requests for the Owners of an object. E.g. the object
	// that created the object that was the source of the Event
	if racr, isDeletePhaseReconciler := r.(*DeletePhaseReconciler); isDeletePhaseReconciler {
		// The enqueueRequestForOwner is not actually done here since we don't know yet the
		// content of the yaml file. The tools wait for the yaml files to be parse. The manager
		// then add the "OwnerReference" to the content of the yaml files. It then invokes the EnqueueRequestForOwner
		owner := av1.NewDeletePhaseVersionKind("", "")
		dependentPredicate := racr.BuildDependentPredicate()
		racr.depResourceWatchUpdater = services.BuildDependentResourceWatchUpdater(mgr, owner, c, *dependentPredicate)
	} else if rrf, isReconcileFunc := r.(*reconcile.Func); isReconcileFunc {
		// Unit test issue
		log.Info("UnitTests", "ReconfileFunc", rrf)
	}

	return nil
}

var _ reconcile.Reconciler = &DeletePhaseReconciler{}

// DeletePhaseReconciler reconciles DeletePhase CRD as K8s SubResources.
type DeletePhaseReconciler struct {
	PhaseReconciler
}

const (
	finalizerDeletePhase = "uninstall-deletephase-resource"
)

// Reconcile reads that state of the cluster for an DeletePhase object and
// makes changes based on the state read and what is in the DeletePhase.Spec
//
// Note: The Controller will requeue the Request to be processed again if the
// returned error is non-nil or Result.Requeue is true, otherwise upon
// completion it will remove the work from the queue.
func (r *DeletePhaseReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reclog := deletephaselog.WithValues("namespace", request.Namespace, "deletephase", request.Name)
	reclog.Info("Reconciling")

	instance := &av1.DeletePhase{}
	instance.SetNamespace(request.Namespace)
	instance.SetName(request.Name)

	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	instance.Init()

	if apierrors.IsNotFound(err) {
		// We are working asynchronously. By the time we receive the event,
		// the object could already be gone
		return reconcile.Result{}, nil
	}

	if err != nil {
		reclog.Error(err, "Failed to lookup DeletePhase")
		return reconcile.Result{}, err
	}

	mgr := r.managerFactory.NewDeletePhaseManager(instance)
	reclog = reclog.WithValues("deletephase", mgr.ResourceName())

	var shouldRequeue bool
	if shouldRequeue, err = r.updateFinalizers(instance); shouldRequeue {
		// Need to requeue because finalizer update does not change metadata.generation
		return reconcile.Result{Requeue: true}, err
	}

	if err := r.ensureSynced(mgr, instance); err != nil {
		if !instance.IsDeleted() {
			// TODO(jeb): Changed the behavior to stop only if we are not
			// in a delete phase.
			return reconcile.Result{}, err
		}
	}

	if instance.IsDeleted() {
		if shouldRequeue, err = r.deleteDeletePhase(mgr, instance); shouldRequeue {
			// Need to requeue because finalizer update does not change metadata.generation
			return reconcile.Result{Requeue: true}, err
		}
		return reconcile.Result{}, err
	}

	if instance.IsTargetStateUninitialized() {
		reclog.Info("TargetState uninitialized; skipping")
		err = r.updateResource(instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = r.client.Status().Update(context.TODO(), instance)
		return reconcile.Result{}, err
	}

	hrc := av1.LcmResourceCondition{
		Type:   av1.ConditionInitialized,
		Status: av1.ConditionStatusTrue,
	}
	instance.Status.SetCondition(hrc, instance.Spec.TargetState)

	switch {
	case !mgr.IsInstalled():
		if shouldRequeue, err = r.installDeletePhase(mgr, instance); shouldRequeue {
			return reconcile.Result{RequeueAfter: r.reconcilePeriod}, err
		}
		return reconcile.Result{}, err
	case mgr.IsUpdateRequired():
		if shouldRequeue, err = r.updateDeletePhase(mgr, instance); shouldRequeue {
			return reconcile.Result{RequeueAfter: r.reconcilePeriod}, err
		}
		return reconcile.Result{}, err
	}

	if err := r.reconcileDeletePhase(mgr, instance); err != nil {
		return reconcile.Result{}, err
	}

	reclog.Info("Reconciled DeletePhase")
	err = r.updateResourceStatus(instance)
	return reconcile.Result{RequeueAfter: r.reconcilePeriod}, err
}

// logAndRecordFailure adds a failure event to the recorder
func (r DeletePhaseReconciler) logAndRecordFailure(instance *av1.DeletePhase, hrc *av1.LcmResourceCondition, err error) {
	reclog := deletephaselog.WithValues("namespace", instance.Namespace, "deletephase", instance.Name)
	reclog.Error(err, fmt.Sprintf("%s. ErrorCondition", hrc.Type.String()))
	r.recorder.Event(instance, corev1.EventTypeWarning, hrc.Type.String(), hrc.Reason.String())
}

// logAndRecordSuccess adds a success event to the recorder
func (r DeletePhaseReconciler) logAndRecordSuccess(instance *av1.DeletePhase, hrc *av1.LcmResourceCondition) {
	reclog := deletephaselog.WithValues("namespace", instance.Namespace, "deletephase", instance.Name)
	reclog.Info(fmt.Sprintf("%s. SuccessCondition", hrc.Type.String()))
	r.recorder.Event(instance, corev1.EventTypeNormal, hrc.Type.String(), hrc.Reason.String())
}

// updateResource updates the Resource object in the cluster
func (r DeletePhaseReconciler) updateResource(instance *av1.DeletePhase) error {
	return r.client.Update(context.TODO(), instance)
}

// updateResourceStatus updates the the Status field of the Resource object in the cluster
func (r DeletePhaseReconciler) updateResourceStatus(instance *av1.DeletePhase) error {
	reclog := deletephaselog.WithValues("namespace", instance.Namespace, "deletephase", instance.Name)

	helper := av1.LcmResourceConditionListHelper{Items: instance.Status.Conditions}
	instance.Status.Conditions = helper.InitIfEmpty()

	// JEB: Be sure to have update status subresources in the CRD.yaml
	// JEB: Look for kubebuilder subresources in the _types.go
	err := r.client.Status().Update(context.TODO(), instance)
	if err != nil {
		reclog.Error(err, "Failure to update status. Ignoring")
		err = nil
	}

	return err
}

// ensureSynced checks that the DeletePhaseManager is in sync with the cluster
func (r DeletePhaseReconciler) ensureSynced(mgr services.DeletePhaseManager, instance *av1.DeletePhase) error {
	if err := mgr.SyncResource(context.TODO()); err != nil {
		hrc := av1.LcmResourceCondition{
			Type:    av1.ConditionIrreconcilable,
			Status:  av1.ConditionStatusTrue,
			Reason:  av1.ReasonReconcileError,
			Message: err.Error(),
		}
		instance.Status.SetCondition(hrc, instance.Spec.TargetState)
		r.logAndRecordFailure(instance, &hrc, err)
		_ = r.updateResourceStatus(instance)
		return err
	}
	instance.Status.RemoveCondition(av1.ConditionIrreconcilable)
	return nil
}

// updateFinalizers asserts that the finalizers match what is expected based on
// whether the instance is currently being deleted or not. It returns true if
// the finalizers were changed, false otherwise
func (r DeletePhaseReconciler) updateFinalizers(instance *av1.DeletePhase) (bool, error) {
	pendingFinalizers := instance.GetFinalizers()
	if !instance.IsDeleted() && !r.contains(pendingFinalizers, finalizerDeletePhase) {
		finalizers := append(pendingFinalizers, finalizerDeletePhase)
		instance.SetFinalizers(finalizers)
		err := r.updateResource(instance)

		return true, err
	}
	return false, nil
}

// watchDependentResources updates all resources which are dependent on this one
func (r DeletePhaseReconciler) watchDependentResources(resource *av1.SubResourceList) error {
	if r.depResourceWatchUpdater != nil {
		if err := r.depResourceWatchUpdater(resource.GetDependentResources()); err != nil {
			return err
		}
	}
	return nil
}

// deleteDeletePhase deletes an instance of an DeletePhase. It returns true if the reconciler should be re-enqueueed
func (r DeletePhaseReconciler) deleteDeletePhase(mgr services.DeletePhaseManager, instance *av1.DeletePhase) (bool, error) {
	reclog := deletephaselog.WithValues("namespace", instance.Namespace, "deletephase", instance.Name)
	reclog.Info("Deleting")

	pendingFinalizers := instance.GetFinalizers()
	if !r.contains(pendingFinalizers, finalizerDeletePhase) {
		reclog.Info("DeletePhase is terminated, skipping reconciliation")
		return false, nil
	}

	uninstalledResource, err := mgr.UninstallResource(context.TODO())
	if err != nil && err != services.ErrNotFound {
		hrc := av1.LcmResourceCondition{
			Type:         av1.ConditionFailed,
			Status:       av1.ConditionStatusTrue,
			Reason:       av1.ReasonUninstallError,
			Message:      err.Error(),
			ResourceName: uninstalledResource.GetName(),
		}
		instance.Status.SetCondition(hrc, instance.Spec.TargetState)
		r.logAndRecordFailure(instance, &hrc, err)

		_ = r.updateResourceStatus(instance)
		return false, err
	}
	instance.Status.RemoveCondition(av1.ConditionFailed)

	if err == services.ErrNotFound {
		reclog.Info("Resource already uninstalled, Removing finalizer")
	} else {
		hrc := av1.LcmResourceCondition{
			Type:   av1.ConditionDeployed,
			Status: av1.ConditionStatusFalse,
			Reason: av1.ReasonUninstallSuccessful,
		}
		instance.Status.SetCondition(hrc, instance.Spec.TargetState)
		r.logAndRecordSuccess(instance, &hrc)
	}
	if err := r.updateResourceStatus(instance); err != nil {
		return false, err
	}

	finalizers := []string{}
	for _, pendingFinalizer := range pendingFinalizers {
		if pendingFinalizer != finalizerDeletePhase {
			finalizers = append(finalizers, pendingFinalizer)
		}
	}
	instance.SetFinalizers(finalizers)
	err = r.updateResource(instance)

	return true, err
}

// installDeletePhase attempts to install instance. It returns true if the reconciler should be re-enqueueed
func (r DeletePhaseReconciler) installDeletePhase(mgr services.DeletePhaseManager, instance *av1.DeletePhase) (bool, error) {
	reclog := deletephaselog.WithValues("namespace", instance.Namespace, "deletephase", instance.Name)
	reclog.Info("Installing")

	installedResource, err := mgr.InstallResource(context.TODO())
	if err != nil {
		instance.Status.RemoveCondition(av1.ConditionRunning)

		hrc := av1.LcmResourceCondition{
			Type:    av1.ConditionFailed,
			Status:  av1.ConditionStatusTrue,
			Reason:  av1.ReasonInstallError,
			Message: err.Error(),
		}
		instance.Status.SetCondition(hrc, instance.Spec.TargetState)
		r.logAndRecordFailure(instance, &hrc, err)

		_ = r.updateResourceStatus(instance)
		return false, err
	}
	instance.Status.RemoveCondition(av1.ConditionFailed)

	if err := r.watchDependentResources(installedResource); err != nil {
		reclog.Error(err, "Failed to update watch on dependent resources")
		return false, err
	}

	hrc := av1.LcmResourceCondition{
		Type:         av1.ConditionRunning,
		Status:       av1.ConditionStatusTrue,
		Reason:       av1.ReasonInstallSuccessful,
		Message:      installedResource.GetPhaseKind().String(),
		ResourceName: installedResource.GetName(),
	}
	instance.Status.SetCondition(hrc, instance.Spec.TargetState)
	r.logAndRecordSuccess(instance, &hrc)

	err = r.updateResourceStatus(instance)
	return true, err
}

// updateDeletePhase attempts to update instance. It returns true if the reconciler should be re-enqueueed
func (r DeletePhaseReconciler) updateDeletePhase(mgr services.DeletePhaseManager, instance *av1.DeletePhase) (bool, error) {
	reclog := deletephaselog.WithValues("namespace", instance.Namespace, "deletephase", instance.Name)
	reclog.Info("Updating")

	previousResource, updatedResource, err := mgr.UpdateResource(context.TODO())
	if previousResource != nil && updatedResource != nil {
		reclog.Info("UpdateResource", "Previous", previousResource.GetName(), "Updated", updatedResource.GetName())
	}
	if err != nil {
		instance.Status.RemoveCondition(av1.ConditionRunning)

		hrc := av1.LcmResourceCondition{
			Type:         av1.ConditionFailed,
			Status:       av1.ConditionStatusTrue,
			Reason:       av1.ReasonUpdateError,
			Message:      err.Error(),
			ResourceName: updatedResource.GetName(),
		}
		instance.Status.SetCondition(hrc, instance.Spec.TargetState)
		r.logAndRecordFailure(instance, &hrc, err)

		_ = r.updateResourceStatus(instance)
		return false, err
	}
	instance.Status.RemoveCondition(av1.ConditionFailed)

	if err := r.watchDependentResources(updatedResource); err != nil {
		reclog.Error(err, "Failed to update watch on dependent resources")
		return false, err
	}

	hrc := av1.LcmResourceCondition{
		Type:         av1.ConditionRunning,
		Status:       av1.ConditionStatusTrue,
		Reason:       av1.ReasonUpdateSuccessful,
		Message:      updatedResource.GetPhaseKind().String(),
		ResourceName: updatedResource.GetName(),
	}
	instance.Status.SetCondition(hrc, instance.Spec.TargetState)
	r.logAndRecordSuccess(instance, &hrc)

	err = r.updateResourceStatus(instance)
	return true, err
}

// reconcileDeletePhase reconciles the phases with the flow
func (r DeletePhaseReconciler) reconcileDeletePhase(mgr services.DeletePhaseManager, instance *av1.DeletePhase) error {
	reclog := deletephaselog.WithValues("namespace", instance.Namespace, "deletephase", instance.Name)
	reclog.Info("Reconciling DeletePhase and LcmResource")

	reconciledResource, err := mgr.ReconcileResource(context.TODO())
	if err != nil {
		instance.Status.RemoveCondition(av1.ConditionRunning)

		hrc := av1.LcmResourceCondition{
			Type:         av1.ConditionIrreconcilable,
			Status:       av1.ConditionStatusTrue,
			Reason:       av1.ReasonReconcileError,
			Message:      err.Error(),
			ResourceName: reconciledResource.GetName(),
		}
		instance.Status.SetCondition(hrc, instance.Spec.TargetState)
		r.logAndRecordFailure(instance, &hrc, err)

		_ = r.updateResourceStatus(instance)
		return err
	}
	instance.Status.RemoveCondition(av1.ConditionIrreconcilable)

	if err := r.watchDependentResources(reconciledResource); err != nil {
		reclog.Error(err, "Failed to update watch on dependent resources")
		return err
	}

	if reconciledResource.IsFailedOrError() {
		// We reconcile. Everything is ready. The flow is now ok
		instance.Status.RemoveCondition(av1.ConditionRunning)

		hrc := av1.LcmResourceCondition{
			Type:         av1.ConditionError,
			Status:       av1.ConditionStatusTrue,
			Reason:       av1.ReasonUnderlyingResourcesError,
			Message:      reconciledResource.GetPhaseKind().String(),
			ResourceName: reconciledResource.GetName(),
		}
		instance.Status.SetCondition(hrc, instance.Spec.TargetState)
		r.logAndRecordSuccess(instance, &hrc)

		err = r.updateResourceStatus(instance)
		return err
	}

	if reconciledResource.IsReady() {
		// We reconcile. Everything is ready. The flow is now ok
		instance.Status.RemoveCondition(av1.ConditionRunning)

		hrc := av1.LcmResourceCondition{
			Type:         av1.ConditionDeployed,
			Status:       av1.ConditionStatusTrue,
			Reason:       av1.ReasonUnderlyingResourcesReady,
			Message:      reconciledResource.GetPhaseKind().String(),
			ResourceName: reconciledResource.GetName(),
		}
		instance.Status.SetCondition(hrc, instance.Spec.TargetState)
		r.logAndRecordSuccess(instance, &hrc)

		err = r.updateResourceStatus(instance)
		return err
	}

	return nil
}
