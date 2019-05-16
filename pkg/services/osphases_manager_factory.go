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

package services

import (
	av1 "github.com/keleustes/oslc-operator/pkg/apis/openstacklcm/v1alpha1"
)

// ManagerFactory creates Managers that are specific to custom resources.
type PhaseManagerFactory interface {
	NewPlanningPhaseManager(r *av1.PlanningPhase) PlanningPhaseManager
	NewInstallPhaseManager(r *av1.InstallPhase) InstallPhaseManager
	NewTestPhaseManager(r *av1.TestPhase) TestPhaseManager
	NewTrafficRolloutPhaseManager(r *av1.TrafficRolloutPhase) TrafficRolloutPhaseManager
	NewOperationalPhaseManager(r *av1.OperationalPhase) OperationalPhaseManager
	NewTrafficDrainPhaseManager(r *av1.TrafficDrainPhase) TrafficDrainPhaseManager
	NewUpgradePhaseManager(r *av1.UpgradePhase) UpgradePhaseManager
	NewRollbackPhaseManager(r *av1.RollbackPhase) RollbackPhaseManager
	NewDeletePhaseManager(r *av1.DeletePhase) DeletePhaseManager
}
