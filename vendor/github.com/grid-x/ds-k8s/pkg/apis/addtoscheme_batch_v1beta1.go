/*
 * Author: Joel Hermanns <j.hermanns@gridx.ai>
 */

package apis

import (
	"github.com/grid-x/ds-k8s/pkg/apis/batch/v1beta1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1beta1.SchemeBuilder.AddToScheme)
}
