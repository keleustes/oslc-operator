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

package oslc

import (
	"context"

	av1 "github.com/keleustes/oslc-operator/pkg/apis/openstacklcm/v1alpha1"
	lcmif "github.com/keleustes/oslc-operator/pkg/services"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type basemanager struct {
	kubeClient     client.Client
	renderer       lcmif.OwnerRefHelmRenderer
	oslcRefs       []metav1.OwnerReference
	oslcName       string
	oslcNamespace  string
	sourceType     string
	sourceLocation string
	serviceName    string

	isInstalled           bool
	isUpdateRequired      bool
	deployedLifecycleFlow *av1.LifecycleFlow
}

// ResourceName returns the name of the release.
func (m basemanager) ResourceName() string {
	return m.oslcName
}

func (m basemanager) IsInstalled() bool {
	return m.isInstalled
}

func (m basemanager) IsUpdateRequired() bool {
	return m.isUpdateRequired
}

// Render a chart or just a file
func (m basemanager) render(ctx context.Context) (*av1.LifecycleFlow, error) {
	var err error
	var subResourceList *av1.SubResourceList

	if m.sourceType == "generate" {
		// In order to use the generic flow, we instantiate on internal chart
		subResourceList, err = m.renderer.RenderChart(m.oslcName, m.oslcNamespace, m.sourceLocation)
	} else if m.sourceType == "tar" {
		subResourceList, err = m.renderer.RenderChart(m.oslcName, m.oslcNamespace, m.sourceLocation)
	} else {
		subResourceList, err = m.renderer.RenderFile(m.oslcName, m.oslcNamespace, m.sourceLocation)
	}

	phaseList := av1.NewLifecycleFlow(m.oslcNamespace, m.oslcName)
	if subResourceList != nil {
		for _, item := range subResourceList.Items {
			if item.GetAPIVersion() == "openstacklcm.airshipit.org/v1alpha1" {
				// TODO(jeb): We should filter on Phase here.
				phaseList.Phases[item.GetKind()] = item
			} else if item.GetAPIVersion() == "argoproj.io/v1alpha1" {
				// TODO(jeb): We should filter on workflow here.
				phaseList.Main = &item
			} else {
				log.Info("Filtering ", "kind", item.GetKind())
			}
		}
	}
	return phaseList, err
}

// SyncResource retrieves from K8s the sub resources (Workflow, Job, ....) attached to this Oslc CR
func (m *basemanager) syncResource(ctx context.Context) error {
	m.deployedLifecycleFlow = av1.NewLifecycleFlow(m.oslcNamespace, m.oslcName)

	rendered, alreadyDeployed, err := m.sync(ctx)
	if err != nil {
		return err
	}

	m.deployedLifecycleFlow = alreadyDeployed
	if len(rendered.GetDependentResources()) != len(alreadyDeployed.GetDependentResources()) {
		m.isInstalled = false
		m.isUpdateRequired = false
	} else {
		m.isInstalled = true
		m.isUpdateRequired = false
	}

	return nil
}

// Attempts to compare the K8s object present with the rendered objects
func (m basemanager) sync(ctx context.Context) (*av1.LifecycleFlow, *av1.LifecycleFlow, error) {
	deployed := av1.NewLifecycleFlow(m.oslcNamespace, m.oslcName)

	rendered, err := m.render(ctx)
	if err != nil {
		return nil, deployed, err
	}

	errs := make([]error, 0)

	if rendered.Main != nil {
		existingResource := unstructured.Unstructured{}
		existingResource.SetAPIVersion(rendered.Main.GetAPIVersion())
		existingResource.SetKind(rendered.Main.GetKind())
		existingResource.SetName(rendered.Main.GetName())
		existingResource.SetNamespace(rendered.Main.GetNamespace())

		err := m.kubeClient.Get(context.TODO(), types.NamespacedName{Name: existingResource.GetName(), Namespace: existingResource.GetNamespace()}, &existingResource)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				log.Error(err, "Can't not retrieve main workflow")
				errs = append(errs, err)
			}
		} else {
			deployed.Main = &existingResource
		}
	}

	for phaseName, renderedResource := range rendered.Phases {
		existingResource := unstructured.Unstructured{}
		existingResource.SetAPIVersion(renderedResource.GetAPIVersion())
		existingResource.SetKind(renderedResource.GetKind())
		existingResource.SetName(renderedResource.GetName())
		existingResource.SetNamespace(renderedResource.GetNamespace())

		err := m.kubeClient.Get(context.TODO(), types.NamespacedName{Name: existingResource.GetName(), Namespace: existingResource.GetNamespace()}, &existingResource)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				log.Error(err, "Can't not retrieve phase")
				errs = append(errs, err)
			}
		} else {
			deployed.Phases[phaseName] = existingResource
		}
	}

	if !deployed.CheckOwnerReference(m.oslcRefs) {
		return rendered, nil, lcmif.OwnershipMismatch
	}

	// TODO(jeb): not sure this is right
	// if len(errs) != 0 {
	//	return rendered, deployed, errs[0]
	// }
	return rendered, deployed, nil
}

// InstallResource creates K8s sub resources (Workflow, Job, ....) attached to this Phase CR
func (m basemanager) installResource(ctx context.Context) (*av1.LifecycleFlow, error) {

	errs := make([]error, 0)
	created := av1.NewLifecycleFlow(m.oslcNamespace, m.oslcName)

	if m.deployedLifecycleFlow == nil {
		// There was an error during SyncResource
		return created, lcmif.InstallError
	}

	rendered, err := m.render(ctx)
	if err != nil {
		return m.deployedLifecycleFlow, err
	}

	for phaseName, toCreate := range rendered.Phases {
		err := m.kubeClient.Create(context.TODO(), &toCreate)
		if err != nil {
			if !apierrors.IsAlreadyExists(err) {
				log.Error(err, "Can't not create Phase")
				errs = append(errs, err)
			} else {
				// Should consider as just created by us
				created.Phases[phaseName] = toCreate
			}
		} else {
			created.Phases[phaseName] = toCreate
		}
	}

	if rendered.Main != nil {
		err := m.kubeClient.Create(context.TODO(), rendered.Main)
		if err != nil {
			if !apierrors.IsAlreadyExists(err) {
				log.Error(err, "Could not create Main Workflow")
				errs = append(errs, err)
			} else {
				// Should consider as just created by us
				created.Main = rendered.Main
			}
		} else {
			created.Main = rendered.Main
		}
	} else {
		log.Info("No Main Workflow")
	}

	if len(errs) != 0 {
		return created, errs[0]
	}
	return created, nil
}

// InstallResource updates K8s sub resources (Workflow, Job, ....) attached to this Phase CR
func (m basemanager) updateResource(ctx context.Context) (*av1.LifecycleFlow, *av1.LifecycleFlow, error) {

	updated := av1.NewLifecycleFlow(m.oslcNamespace, m.oslcName)

	if m.deployedLifecycleFlow == nil {
		// There was an error during SyncResource
		return m.deployedLifecycleFlow, updated, lcmif.UpdateError
	}

	// TODO(JEB): Big hack. ReconcileResource should do more
	m.deployedLifecycleFlow.DeepCopyInto(updated)

	return m.deployedLifecycleFlow, updated, nil
}

// ReconcileResource creates or patches resources as necessary to match this Phase CR
func (m basemanager) reconcileResource(ctx context.Context) (*av1.LifecycleFlow, error) {

	reconciled := av1.NewLifecycleFlow(m.oslcNamespace, m.oslcName)

	if m.deployedLifecycleFlow == nil {
		return reconciled, lcmif.ReconcileError
	}

	// TODO(JEB): Big hack. ReconcileResource should do more
	m.deployedLifecycleFlow.DeepCopyInto(reconciled)

	return reconciled, nil
}

// UninstallResource delete K8s sub resources (Workflow, Job, ....) attached to this Phase CR
func (m basemanager) uninstallResource(ctx context.Context) (*av1.LifecycleFlow, error) {
	errs := make([]error, 0)
	notdeleted := av1.NewLifecycleFlow(m.oslcNamespace, m.oslcName)

	if m.deployedLifecycleFlow == nil {
		// There was an error during SyncResource
		return notdeleted, lcmif.UninstallError
	}

	if m.deployedLifecycleFlow.Main != nil {
		err := m.kubeClient.Delete(context.TODO(), m.deployedLifecycleFlow.Main)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				log.Error(err, "Can't not delete main flow")
				errs = append(errs, err)
				notdeleted.Main = m.deployedLifecycleFlow.Main
			}
		}
	}

	for phaseName, toDelete := range m.deployedLifecycleFlow.Phases {
		err := m.kubeClient.Delete(context.TODO(), &toDelete)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				log.Error(err, "Can't not delete phase")
				errs = append(errs, err)
				notdeleted.Phases[phaseName] = toDelete
			}
		}
	}

	if len(errs) != 0 {
		return notdeleted, errs[0]
	}
	return notdeleted, nil
}
