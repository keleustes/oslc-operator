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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	av1 "github.com/keleustes/oslc-operator/pkg/apis/openstacklcm/v1alpha1"
	lcmif "github.com/keleustes/oslc-operator/pkg/services"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type managerFactory struct {
	kubeClient client.Client
}

// Simple function to init the renderFiles passed to the helm renderer
func initRenderFiles(stage av1.OslcPhase) []string {
	renderFiles := make([]string, 0)
	return renderFiles
}

// Simple function to init the renderValues passed to the helm renderer
func initRenderValues(stage av1.OslcPhase) map[string]interface{} {
	oslcValues := map[string]interface{}{}
	oslcValues["stage"] = stage.String()
	renderValues := map[string]interface{}{}
	renderValues["oslc"] = oslcValues
	renderValues["lifecycle"] = stage.String()
	return renderValues
}

// NewManagerFactory returns a new factory.
func NewManagerFactory(mgr manager.Manager) lcmif.PhaseManagerFactory {
	return &managerFactory{kubeClient: mgr.GetClient()}
}

// NewPlanningPhaseManager returns a new manager capable of controlling PlanningPhase phase of the service lifecyle
func (f managerFactory) NewPlanningPhaseManager(r *av1.PlanningPhase) lcmif.PlanningPhaseManager {
	controllerRef := metav1.NewControllerRef(r, r.GroupVersionKind())
	ownerRefs := []metav1.OwnerReference{
		*controllerRef,
	}

	renderFiles := initRenderFiles(av1.PhasePlanning)
	renderValues := initRenderValues(av1.PhasePlanning)
	renderer := &planningrenderer{
		helmrenderer: NewOwnerRefHelmRenderer(ownerRefs, "osplan", renderFiles, renderValues),
		spec:         r.Spec,
	}

	return &planningmanager{
		phasemanager: phasemanager{
			kubeClient:     f.kubeClient,
			renderer:       renderer,
			source:         r.Spec.Source,
			serviceName:    r.Spec.OpenstackServiceName,
			phaseRefs:      ownerRefs,
			phaseName:      r.GetName(),
			phaseNamespace: r.GetNamespace()},

		spec:   r.Spec,
		status: &r.Status,
	}
}

// NewInstallPhaseManager returns a new manager capable of controlling InstallPhase phase of the service lifecyle
func (f managerFactory) NewInstallPhaseManager(r *av1.InstallPhase) lcmif.InstallPhaseManager {
	controllerRef := metav1.NewControllerRef(r, r.GroupVersionKind())
	ownerRefs := []metav1.OwnerReference{
		*controllerRef,
	}

	renderFiles := initRenderFiles(av1.PhaseInstall)
	renderValues := initRenderValues(av1.PhaseInstall)
	renderer := &installrenderer{
		helmrenderer: NewOwnerRefHelmRenderer(ownerRefs, "osins", renderFiles, renderValues),
		spec:         r.Spec,
	}

	return &installmanager{
		phasemanager: phasemanager{
			kubeClient:     f.kubeClient,
			renderer:       renderer,
			source:         r.Spec.Source,
			serviceName:    r.Spec.OpenstackServiceName,
			phaseRefs:      ownerRefs,
			phaseName:      r.GetName(),
			phaseNamespace: r.GetNamespace()},

		spec:   r.Spec,
		status: &r.Status,
	}
}

// NewTestPhaseManager returns a new manager capable of controlling TestPhase phase of the service lifecyle
func (f managerFactory) NewTestPhaseManager(r *av1.TestPhase) lcmif.TestPhaseManager {
	controllerRef := metav1.NewControllerRef(r, r.GroupVersionKind())
	ownerRefs := []metav1.OwnerReference{
		*controllerRef,
	}

	renderFiles := initRenderFiles(av1.PhaseTest)
	renderValues := initRenderValues(av1.PhaseTest)
	renderer := &testrenderer{
		helmrenderer: NewOwnerRefHelmRenderer(ownerRefs, "ostest", renderFiles, renderValues),
		spec:         r.Spec,
	}

	return &testmanager{
		phasemanager: phasemanager{
			kubeClient:     f.kubeClient,
			renderer:       renderer,
			source:         r.Spec.Source,
			serviceName:    r.Spec.OpenstackServiceName,
			phaseRefs:      ownerRefs,
			phaseName:      r.GetName(),
			phaseNamespace: r.GetNamespace()},

		spec:   r.Spec,
		status: &r.Status,
	}
}

// NewTrafficRolloutPhaseManager returns a new manager capable of controlling TrafficRolloutPhase phase of the service lifecyle
func (f managerFactory) NewTrafficRolloutPhaseManager(r *av1.TrafficRolloutPhase) lcmif.TrafficRolloutPhaseManager {
	controllerRef := metav1.NewControllerRef(r, r.GroupVersionKind())
	ownerRefs := []metav1.OwnerReference{
		*controllerRef,
	}

	renderFiles := initRenderFiles(av1.PhaseTrafficRollout)
	renderValues := initRenderValues(av1.PhaseTrafficRollout)
	renderer := &trafficrolloutrenderer{
		helmrenderer: NewOwnerRefHelmRenderer(ownerRefs, "osroll", renderFiles, renderValues),
		spec:         r.Spec,
	}

	return &trafficrolloutmanager{
		phasemanager: phasemanager{
			kubeClient:     f.kubeClient,
			renderer:       renderer,
			source:         r.Spec.Source,
			serviceName:    r.Spec.OpenstackServiceName,
			phaseRefs:      ownerRefs,
			phaseName:      r.GetName(),
			phaseNamespace: r.GetNamespace()},

		spec:   r.Spec,
		status: &r.Status,
	}
}

// NewOperationalPhaseManager returns a new manager capable of controlling OperationalPhase phase of the service lifecyle
func (f managerFactory) NewOperationalPhaseManager(r *av1.OperationalPhase) lcmif.OperationalPhaseManager {
	controllerRef := metav1.NewControllerRef(r, r.GroupVersionKind())
	ownerRefs := []metav1.OwnerReference{
		*controllerRef,
	}

	renderFiles := initRenderFiles(av1.PhaseOperational)
	renderValues := initRenderValues(av1.PhaseOperational)
	renderer := &operationalrenderer{
		helmrenderer: NewOwnerRefHelmRenderer(ownerRefs, "osops", renderFiles, renderValues),
		spec:         r.Spec,
	}

	return &operationalmanager{
		phasemanager: phasemanager{
			kubeClient:     f.kubeClient,
			renderer:       renderer,
			source:         r.Spec.Source,
			serviceName:    r.Spec.OpenstackServiceName,
			phaseRefs:      ownerRefs,
			phaseName:      r.GetName(),
			phaseNamespace: r.GetNamespace()},

		spec:   r.Spec,
		status: &r.Status,
	}
}

// NewTrafficDrainPhaseManager returns a new manager capable of controlling TrafficDrainPhase phase of the service lifecyle
func (f managerFactory) NewTrafficDrainPhaseManager(r *av1.TrafficDrainPhase) lcmif.TrafficDrainPhaseManager {
	controllerRef := metav1.NewControllerRef(r, r.GroupVersionKind())
	ownerRefs := []metav1.OwnerReference{
		*controllerRef,
	}

	renderFiles := initRenderFiles(av1.PhaseTrafficDrain)
	renderValues := initRenderValues(av1.PhaseTrafficDrain)
	renderer := &trafficdrainrenderer{
		helmrenderer: NewOwnerRefHelmRenderer(ownerRefs, "osdrain", renderFiles, renderValues),
		spec:         r.Spec,
	}

	return &trafficdrainmanager{
		phasemanager: phasemanager{
			kubeClient:     f.kubeClient,
			renderer:       renderer,
			source:         r.Spec.Source,
			serviceName:    r.Spec.OpenstackServiceName,
			phaseRefs:      ownerRefs,
			phaseName:      r.GetName(),
			phaseNamespace: r.GetNamespace()},

		spec:   r.Spec,
		status: &r.Status,
	}
}

// NewUpgradePhaseManager returns a new manager capable of controlling UpgradePhase phase of the service lifecyle
func (f managerFactory) NewUpgradePhaseManager(r *av1.UpgradePhase) lcmif.UpgradePhaseManager {
	controllerRef := metav1.NewControllerRef(r, r.GroupVersionKind())
	ownerRefs := []metav1.OwnerReference{
		*controllerRef,
	}

	renderFiles := initRenderFiles(av1.PhaseUpgrade)
	renderValues := initRenderValues(av1.PhaseUpgrade)
	renderer := &upgraderenderer{
		helmrenderer: NewOwnerRefHelmRenderer(ownerRefs, "osupg", renderFiles, renderValues),
		spec:         r.Spec,
	}

	return &upgrademanager{
		phasemanager: phasemanager{
			kubeClient:     f.kubeClient,
			renderer:       renderer,
			source:         r.Spec.Source,
			serviceName:    r.Spec.OpenstackServiceName,
			phaseRefs:      ownerRefs,
			phaseName:      r.GetName(),
			phaseNamespace: r.GetNamespace()},

		spec:   r.Spec,
		status: &r.Status,
	}
}

// NewRollbackPhaseManager returns a new manager capable of controlling RollbackPhase phase of the service lifecyle
func (f managerFactory) NewRollbackPhaseManager(r *av1.RollbackPhase) lcmif.RollbackPhaseManager {
	controllerRef := metav1.NewControllerRef(r, r.GroupVersionKind())
	ownerRefs := []metav1.OwnerReference{
		*controllerRef,
	}

	renderFiles := initRenderFiles(av1.PhaseRollback)
	renderValues := initRenderValues(av1.PhaseRollback)
	renderer := &rollbackrenderer{
		helmrenderer: NewOwnerRefHelmRenderer(ownerRefs, "osrbck", renderFiles, renderValues),
		spec:         r.Spec,
	}

	return &rollbackmanager{
		phasemanager: phasemanager{
			kubeClient:     f.kubeClient,
			renderer:       renderer,
			source:         r.Spec.Source,
			serviceName:    r.Spec.OpenstackServiceName,
			phaseRefs:      ownerRefs,
			phaseName:      r.GetName(),
			phaseNamespace: r.GetNamespace()},

		spec:   r.Spec,
		status: &r.Status,
	}
}

// NewDeletePhaseManager returns a new manager capable of controlling DeletePhase phase of the service lifecyle
func (f managerFactory) NewDeletePhaseManager(r *av1.DeletePhase) lcmif.DeletePhaseManager {
	controllerRef := metav1.NewControllerRef(r, r.GroupVersionKind())
	ownerRefs := []metav1.OwnerReference{
		*controllerRef,
	}

	renderFiles := initRenderFiles(av1.PhaseDelete)
	renderValues := initRenderValues(av1.PhaseDelete)
	renderer := &deleterenderer{
		helmrenderer: NewOwnerRefHelmRenderer(ownerRefs, "osdlt", renderFiles, renderValues),
		spec:         r.Spec,
	}

	return &deletemanager{
		phasemanager: phasemanager{
			kubeClient:     f.kubeClient,
			renderer:       renderer,
			source:         r.Spec.Source,
			serviceName:    r.Spec.OpenstackServiceName,
			phaseRefs:      ownerRefs,
			phaseName:      r.GetName(),
			phaseNamespace: r.GetNamespace()},

		spec:   r.Spec,
		status: &r.Status,
	}
}
