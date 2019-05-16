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

	av1 "github.com/keleustes/oslc-operator/pkg/apis/openstacklcm/v1alpha1"
	lcmif "github.com/keleustes/oslc-operator/pkg/services"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type phasemanager struct {
	kubeClient     client.Client
	renderer       lcmif.OwnerRefHelmRenderer
	serviceName    string
	phaseRefs      []metav1.OwnerReference
	phaseName      string
	phaseNamespace string
	source         *av1.PhaseSource

	isInstalled             bool
	isUpdateRequired        bool
	deployedSubResourceList *av1.SubResourceList
}

// ResourceName returns the name of the release.
func (m phasemanager) ResourceName() string {
	return m.phaseName
}

func (m phasemanager) IsInstalled() bool {
	return m.isInstalled
}

func (m phasemanager) IsUpdateRequired() bool {
	return m.isUpdateRequired
}

// Render a chart or just a file
func (m phasemanager) render(ctx context.Context) (*av1.SubResourceList, error) {
	if m.source.Type == "tar" {
		return m.renderer.RenderChart(m.phaseName, m.phaseNamespace, m.source.Location)
	} else {
		return m.renderer.RenderFile(m.phaseName, m.phaseNamespace, m.source.Location)
	}
}

// Try to compare the resource in the CRD and the resources in Kubernetes
func (m *phasemanager) syncResource(ctx context.Context) error {

	m.deployedSubResourceList = av1.NewSubResourceList(m.phaseNamespace, m.phaseName)

	rendered, deployed, err := m.sync(ctx)
	if err != nil {
		return err
	}

	m.deployedSubResourceList = deployed
	if len(rendered.Items) != len(deployed.Items) {
		m.isInstalled = false
		m.isUpdateRequired = false
	} else {
		m.isInstalled = true
		m.isUpdateRequired = false
	}

	return nil
}

// Attempts to compare the K8s object present with the rendered objects
func (m phasemanager) sync(ctx context.Context) (*av1.SubResourceList, *av1.SubResourceList, error) {
	alreadyDeployed := av1.NewSubResourceList(m.phaseNamespace, m.phaseName)

	rendered, err := m.render(ctx)
	if err != nil {
		return nil, alreadyDeployed, err
	}

	errs := make([]error, 0)
	for _, renderedResource := range rendered.Items {
		// TODO(jeb): Don't undestand why need to code such a klduge
		existingResource := unstructured.Unstructured{}
		existingResource.SetAPIVersion(renderedResource.GetAPIVersion())
		existingResource.SetKind(renderedResource.GetKind())
		existingResource.SetName(renderedResource.GetName())
		existingResource.SetNamespace(renderedResource.GetNamespace())

		err := m.kubeClient.Get(context.TODO(), types.NamespacedName{Name: existingResource.GetName(), Namespace: existingResource.GetNamespace()}, &existingResource)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				// Don't want to trace is the error is not a NotFound.
				log.Error(err, "Can't not retrieve Resource")
				errs = append(errs, err)
			}
		} else {
			alreadyDeployed.Items = append(alreadyDeployed.Items, existingResource)
		}
	}

	if !alreadyDeployed.CheckOwnerReference(m.phaseRefs) {
		return rendered, nil, lcmif.OwnershipMismatch
	}

	// TODO(jeb): not sure this is right
	// if len(errs) != 0 {
	//	return rendered, alreadyDeployed, errs[0]
	// }
	return rendered, alreadyDeployed, nil
}

// InstallResource creates K8s sub resources (Workflow, Job, ....) attached to this Phase CR
func (m phasemanager) installResource(ctx context.Context) (*av1.SubResourceList, error) {

	errs := make([]error, 0)
	created := av1.NewSubResourceList(m.phaseNamespace, m.phaseName)

	if m.deployedSubResourceList == nil {
		// There was an error during SyncResource
		return created, lcmif.InstallError
	}

	rendered, err := m.render(ctx)
	if err != nil {
		return created, err
	}

	rendered.Items = lcmif.SortByInstallOrder(rendered.Items)
	for _, toCreate := range rendered.Items {
		err := m.kubeClient.Create(context.TODO(), &toCreate)
		if err != nil {
			if !apierrors.IsAlreadyExists(err) {
				log.Error(err, "Can't not Create Resource", "kind", toCreate.GetKind(), "name", toCreate.GetName())
				errs = append(errs, err)
			} else {
				// Should consider as just created by us ?
				// created.Items = append(created.Items, toCreate)
			}
		} else {
			log.Info("Created Resource", "kind", toCreate.GetKind(), "name", toCreate.GetName())
			created.Items = append(created.Items, toCreate)
		}
	}

	if len(errs) != 0 {
		return created, errs[0]
	}
	return created, nil
}

// InstallResource updates K8s sub resources (Workflow, Job, ....) attached to this Phase CR
func (m phasemanager) updateResource(ctx context.Context) (*av1.SubResourceList, *av1.SubResourceList, error) {

	updated := av1.NewSubResourceList(m.phaseNamespace, m.phaseName)

	if m.deployedSubResourceList == nil {
		// There was an error during SyncResource
		return m.deployedSubResourceList, updated, lcmif.UpdateError
	}

	// TODO(JEB): Big hack. ReconcileResource should do more
	m.deployedSubResourceList.DeepCopyInto(updated)

	return m.deployedSubResourceList, updated, nil
}

// ReconcileResource creates or patches resources as necessary to match this Phase CR
func (m phasemanager) reconcileResource(ctx context.Context) (*av1.SubResourceList, error) {

	reconciled := av1.NewSubResourceList(m.phaseNamespace, m.phaseName)

	if m.deployedSubResourceList == nil {
		// There was an error during SyncResource
		return reconciled, lcmif.ReconcileError
	}

	// TODO(JEB): Big hack. ReconcileResource should do more
	m.deployedSubResourceList.DeepCopyInto(reconciled)

	return reconciled, nil
}

// UninstallResource delete K8s sub resources (Workflow, Job, ....) attached to this Phase CR
func (m phasemanager) uninstallResource(ctx context.Context) (*av1.SubResourceList, error) {
	errs := make([]error, 0)
	notdeleted := av1.NewSubResourceList(m.phaseNamespace, m.phaseName)

	if m.deployedSubResourceList == nil {
		// There was an error during SyncResource
		return notdeleted, lcmif.UninstallError
	}

	m.deployedSubResourceList.Items = lcmif.SortByUninstallOrder(m.deployedSubResourceList.Items)
	for _, toDelete := range m.deployedSubResourceList.Items {
		err := m.kubeClient.Delete(context.TODO(), &toDelete)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				log.Error(err, "Can't not delete Resource")
				errs = append(errs, err)
				notdeleted.Items = append(notdeleted.Items, toDelete)
			}
		}
	}

	if len(errs) != 0 {
		return notdeleted, errs[0]
	}
	return notdeleted, nil
}
