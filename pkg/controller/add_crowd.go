package controller

import (
	"github.com/example-inc/app-operator/pkg/controller/crowd"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, crowd.Add)
}
