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
)

type trafficrolloutmanager struct {
	phasemanager

	spec   av1.TrafficRolloutPhaseSpec
	status *av1.TrafficRolloutPhaseStatus
}

type trafficrolloutrenderer struct {
	helmrenderer lcmif.OwnerRefHelmRenderer

	spec av1.TrafficRolloutPhaseSpec
}

// RenderFile injects TrafficRolloutPhase spec into the rendering of a file
func (o trafficrolloutrenderer) RenderFile(name string, namespace string, fileName string) (*av1.SubResourceList, error) {
	return o.helmrenderer.RenderFile(name, namespace, fileName)
}

// RenderChart injects TrafficRolloutPhase spec into the renderering of a chart
func (o trafficrolloutrenderer) RenderChart(name string, namespace string, chartLocation string) (*av1.SubResourceList, error) {
	return o.helmrenderer.RenderChart(name, namespace, chartLocation)
}

// SyncResource retrieves from K8s the sub resources (Workflow, Job, ....) attached to this TrafficRolloutPhase CR
func (m *trafficrolloutmanager) SyncResource(ctx context.Context) error {
	return m.syncResource(ctx)
}

// InstallResource creates K8s sub resources (Workflow, Job, ....) attached to this TrafficRolloutPhase CR
func (m trafficrolloutmanager) InstallResource(ctx context.Context) (*av1.SubResourceList, error) {
	return m.installResource(ctx)
}

// InstallResource updates K8s sub resources (Workflow, Job, ....) attached to this TrafficRolloutPhase CR
func (m trafficrolloutmanager) UpdateResource(ctx context.Context) (*av1.SubResourceList, *av1.SubResourceList, error) {
	return m.updateResource(ctx)
}

// ReconcileResource creates or patches resources as necessary to match this TrafficRolloutPhase CR
func (m trafficrolloutmanager) ReconcileResource(ctx context.Context) (*av1.SubResourceList, error) {
	return m.reconcileResource(ctx)
}

// UninstallResource trafficrollout K8s sub resources (Workflow, Job, ....) attached to this TrafficRolloutPhase CR
func (m trafficrolloutmanager) UninstallResource(ctx context.Context) (*av1.SubResourceList, error) {
	return m.uninstallResource(ctx)
}
