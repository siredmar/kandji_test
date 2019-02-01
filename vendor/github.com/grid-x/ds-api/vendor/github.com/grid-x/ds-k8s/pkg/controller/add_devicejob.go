/*
 * Author: Joel Hermanns <j.hermanns@gridx.ai>
 */

package controller

import (
	"github.com/grid-x/ds-k8s/pkg/controller/devicejob"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, devicejob.Add)
}
