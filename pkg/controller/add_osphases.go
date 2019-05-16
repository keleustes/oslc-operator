package controller

import (
	"github.com/keleustes/oslc-operator/pkg/controller/osphases"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, osphases.AddPlanningPhaseController)
	AddToManagerFuncs = append(AddToManagerFuncs, osphases.AddInstallPhaseController)
	AddToManagerFuncs = append(AddToManagerFuncs, osphases.AddTestPhaseController)
	AddToManagerFuncs = append(AddToManagerFuncs, osphases.AddTrafficRolloutPhaseController)
	AddToManagerFuncs = append(AddToManagerFuncs, osphases.AddOperationalPhaseController)
	AddToManagerFuncs = append(AddToManagerFuncs, osphases.AddTrafficDrainPhaseController)
	AddToManagerFuncs = append(AddToManagerFuncs, osphases.AddUpgradePhaseController)
	AddToManagerFuncs = append(AddToManagerFuncs, osphases.AddRollbackPhaseController)
	AddToManagerFuncs = append(AddToManagerFuncs, osphases.AddDeletePhaseController)
}
