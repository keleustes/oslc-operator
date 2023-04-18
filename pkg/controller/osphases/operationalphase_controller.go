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

	av1 "github.com/keleustes/armada-crd/pkg/apis/openstacklcm/v1alpha1"
	operationalphasemgr "github.com/keleustes/oslc-operator/pkg/osphases"
	services "github.com/keleustes/oslc-operator/pkg/services"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	crthandler "sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var operationalphaselog = logf.Log.WithName("operationalphase-controller")

// AddOperationalPhaseController creates a new OperationalPhase Controller and adds it to
// the Manager. The Manager will set fields on the Controller and Start it when
// the Manager is Started.
func AddOperationalPhaseController(mgr manager.Manager) error {
	return addOperationalPhase(mgr, newOperationalPhaseReconciler(mgr))
}

// newOperationalPhaseReconciler returns a new reconcile.Reconciler
func newOperationalPhaseReconciler(mgr manager.Manager) reconcile.Reconciler {
	r := &OperationalPhaseReconciler{
		PhaseReconciler: PhaseReconciler{
			client:         mgr.GetClient(),
			scheme:         mgr.GetScheme(),
			recorder:       mgr.GetEventRecorderFor("operationalphase-recorder"),
			managerFactory: operationalphasemgr.NewManagerFactory(mgr),
			// reconcilePeriod: flags.ReconcilePeriod,
		},
	}
	return r
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func addOperationalPhase(mgr manager.Manager, r reconcile.Reconciler) error {

	// Create a new controller
	c, err := controller.New("operationalphase-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource OperationalPhase
	// EnqueueRequestForObject enqueues a Request containing the Name and Namespace of the object
	// that is the source of the Event. (e.g. the created / deleted / updated objects Name and Namespace).
	err = c.Watch(&source.Kind{Type: &av1.OperationalPhase{}}, &crthandler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource (described in the yaml file/chart) and requeue the owner OperationalPhase
	// EnqueueRequestForOwner enqueues Requests for the Owners of an object. E.g. the object
	// that created the object that was the source of the Event
	if racr, isOperationalPhaseReconciler := r.(*OperationalPhaseReconciler); isOperationalPhaseReconciler {
		// The enqueueRequestForOwner is not actually done here since we don't know yet the
		// content of the yaml file. The tools wait for the yaml files to be parse. The manager
		// then add the "OwnerReference" to the content of the yaml files. It then invokes the EnqueueRequestForOwner
		owner := av1.NewOperationalPhaseVersionKind("", "")
		dependentPredicate := racr.BuildDependentPredicate()
		racr.depResourceWatchUpdater = services.BuildDependentResourceWatchUpdater(mgr, owner, c, *dependentPredicate)
	} else if rrf, isReconcileFunc := r.(*reconcile.Func); isReconcileFunc {
		// Unit test issue
		log.Info("UnitTests", "ReconfileFunc", rrf)
	}

	return nil
}

var _ reconcile.Reconciler = &OperationalPhaseReconciler{}

// OperationalPhaseReconciler reconciles OperationalPhase CRD as K8s SubResources.
type OperationalPhaseReconciler struct {
	PhaseReconciler
}

const (
	finalizerOperationalPhase = "uninstall-operationalphase-resource"
)

// Reconcile reads that state of the cluster for an OperationalPhase object and
// makes changes based on the state read and what is in the OperationalPhase.Spec
//
// Note: The Controller will requeue the Request to be processed again if the
// returned error is non-nil or Result.Requeue is true, otherwise upon
// completion it will remove the work from the queue.
func (r *OperationalPhaseReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	reclog := operationalphaselog.WithValues("namespace", request.Namespace, "operationalphase", request.Name)
	reclog.Info("Reconciling")

	instance := &av1.OperationalPhase{}
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
		reclog.Error(err, "Failed to lookup OperationalPhase")
		return reconcile.Result{}, err
	}

	mgr := r.managerFactory.NewOperationalPhaseManager(instance)
	reclog = reclog.WithValues("operationalphase", mgr.ResourceName())

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
		if shouldRequeue, err = r.deleteOperationalPhase(mgr, instance); shouldRequeue {
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
		if shouldRequeue, err = r.installOperationalPhase(mgr, instance); shouldRequeue {
			return reconcile.Result{RequeueAfter: r.reconcilePeriod}, err
		}
		return reconcile.Result{}, err
	case mgr.IsUpdateRequired():
		if shouldRequeue, err = r.updateOperationalPhase(mgr, instance); shouldRequeue {
			return reconcile.Result{RequeueAfter: r.reconcilePeriod}, err
		}
		return reconcile.Result{}, err
	}

	if err := r.reconcileOperationalPhase(mgr, instance); err != nil {
		return reconcile.Result{}, err
	}

	reclog.Info("Reconciled OperationalPhase")
	err = r.updateResourceStatus(instance)
	return reconcile.Result{RequeueAfter: r.reconcilePeriod}, err
}

// logAndRecordFailure adds a failure event to the recorder
func (r OperationalPhaseReconciler) logAndRecordFailure(instance *av1.OperationalPhase, hrc *av1.LcmResourceCondition, err error) {
	reclog := operationalphaselog.WithValues("namespace", instance.Namespace, "operationalphase", instance.Name)
	reclog.Error(err, fmt.Sprintf("%s. ErrorCondition", hrc.Type.String()))
	r.recorder.Event(instance, corev1.EventTypeWarning, hrc.Type.String(), hrc.Reason.String())
}

// logAndRecordSuccess adds a success event to the recorder
func (r OperationalPhaseReconciler) logAndRecordSuccess(instance *av1.OperationalPhase, hrc *av1.LcmResourceCondition) {
	reclog := operationalphaselog.WithValues("namespace", instance.Namespace, "operationalphase", instance.Name)
	reclog.Info(fmt.Sprintf("%s. SuccessCondition", hrc.Type.String()))
	r.recorder.Event(instance, corev1.EventTypeNormal, hrc.Type.String(), hrc.Reason.String())
}

// updateResource updates the Resource object in the cluster
func (r OperationalPhaseReconciler) updateResource(instance *av1.OperationalPhase) error {
	return r.client.Update(context.TODO(), instance)
}

// updateResourceStatus updates the the Status field of the Resource object in the cluster
func (r OperationalPhaseReconciler) updateResourceStatus(instance *av1.OperationalPhase) error {
	reclog := operationalphaselog.WithValues("namespace", instance.Namespace, "operationalphase", instance.Name)

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

// ensureSynced checks that the OperationalPhaseManager is in sync with the cluster
func (r OperationalPhaseReconciler) ensureSynced(mgr services.OperationalPhaseManager, instance *av1.OperationalPhase) error {
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
func (r OperationalPhaseReconciler) updateFinalizers(instance *av1.OperationalPhase) (bool, error) {
	pendingFinalizers := instance.GetFinalizers()
	if !instance.IsDeleted() && !r.contains(pendingFinalizers, finalizerOperationalPhase) {
		finalizers := append(pendingFinalizers, finalizerOperationalPhase)
		instance.SetFinalizers(finalizers)
		err := r.updateResource(instance)

		return true, err
	}
	return false, nil
}

// watchDependentResources updates all resources which are dependent on this one
func (r OperationalPhaseReconciler) watchDependentResources(resource *av1.SubResourceList) error {
	if r.depResourceWatchUpdater != nil {
		if err := r.depResourceWatchUpdater(resource.GetDependentResources()); err != nil {
			return err
		}
	}
	return nil
}

// deleteOperationalPhase deletes an instance of an OperationalPhase. It returns true if the reconciler should be re-enqueueed
func (r OperationalPhaseReconciler) deleteOperationalPhase(mgr services.OperationalPhaseManager, instance *av1.OperationalPhase) (bool, error) {
	reclog := operationalphaselog.WithValues("namespace", instance.Namespace, "operationalphase", instance.Name)
	reclog.Info("Deleting")

	pendingFinalizers := instance.GetFinalizers()
	if !r.contains(pendingFinalizers, finalizerOperationalPhase) {
		reclog.Info("OperationalPhase is terminated, skipping reconciliation")
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
		if pendingFinalizer != finalizerOperationalPhase {
			finalizers = append(finalizers, pendingFinalizer)
		}
	}
	instance.SetFinalizers(finalizers)
	err = r.updateResource(instance)

	return true, err
}

// installOperationalPhase attempts to install instance. It returns true if the reconciler should be re-enqueueed
func (r OperationalPhaseReconciler) installOperationalPhase(mgr services.OperationalPhaseManager, instance *av1.OperationalPhase) (bool, error) {
	reclog := operationalphaselog.WithValues("namespace", instance.Namespace, "operationalphase", instance.Name)
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

// updateOperationalPhase attempts to update instance. It returns true if the reconciler should be re-enqueueed
func (r OperationalPhaseReconciler) updateOperationalPhase(mgr services.OperationalPhaseManager, instance *av1.OperationalPhase) (bool, error) {
	reclog := operationalphaselog.WithValues("namespace", instance.Namespace, "operationalphase", instance.Name)
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

// reconcileOperationalPhase reconciles the phases with the flow
func (r OperationalPhaseReconciler) reconcileOperationalPhase(mgr services.OperationalPhaseManager, instance *av1.OperationalPhase) error {
	reclog := operationalphaselog.WithValues("namespace", instance.Namespace, "operationalphase", instance.Name)
	reclog.Info("Reconciling OperationalPhase and LcmResource")

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
