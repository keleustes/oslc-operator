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
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	av1 "github.com/keleustes/armada-crd/pkg/apis/openstacklcm/v1alpha1"
	lcmif "github.com/keleustes/oslc-operator/pkg/services"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type managerFactory struct {
	kubeClient client.Client
}

// NewManagerFactory returns a new factory.
func NewManagerFactory(mgr manager.Manager) lcmif.OslcManagerFactory {
	return &managerFactory{kubeClient: mgr.GetClient()}
}

// Simple function to init the renderFiles passed to the helm renderer
func initRenderFiles(stage av1.OslcFlowKind) []string {
	renderFiles := make([]string, 0)
	return renderFiles
}

// Simple function to init the renderValues passed to the helm renderer
func initRenderValues(stage av1.OslcFlowKind) map[string]interface{} {
	oslcValues := map[string]interface{}{}
	oslcValues["flow_kind"] = stage.String()
	renderValues := map[string]interface{}{}
	renderValues["oslc"] = oslcValues
	return renderValues
}

// NewOslcManager returns a new manager capable of controlling Oslc phase of the service lifecyle
func (f managerFactory) NewOslcManager(r *av1.Oslc) lcmif.OslcManager {
	controllerRef := metav1.NewControllerRef(r, r.GroupVersionKind())
	ownerRefs := []metav1.OwnerReference{
		*controllerRef,
	}

	renderFiles := initRenderFiles(r.Spec.FlowKind)
	renderValues := initRenderValues(r.Spec.FlowKind)

	sourceType := r.Spec.Source.Type
	sourceLocation := r.Spec.Source.Location
	serviceName := r.Spec.ServiceName

	// TODO(jeb): Kind of kludgy
	if sourceType == "generate" {
		sourceLocation = strings.ReplaceAll(r.Spec.Source.Location, r.Spec.ServiceName, lcmif.GenericFlows.String())
		renderValues["serviceName"] = serviceName
	}

	renderer := &oslcrenderer{
		helmrenderer: NewOwnerRefHelmRenderer(ownerRefs, "oslc", renderFiles, renderValues),
		spec:         r.Spec,
	}

	return &oslcmanager{
		basemanager: basemanager{
			kubeClient:     f.kubeClient,
			renderer:       renderer,
			serviceName:    serviceName,
			sourceType:     sourceType,
			sourceLocation: sourceLocation,
			oslcRefs:       ownerRefs,
			oslcName:       r.GetName(),
			oslcNamespace:  r.GetNamespace()},

		spec:   r.Spec,
		status: &r.Status,
	}
}
