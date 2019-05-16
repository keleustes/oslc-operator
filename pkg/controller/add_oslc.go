package controller

import (
	"github.com/keleustes/oslc-operator/pkg/controller/oslc"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, oslc.AddOslcController)
}
