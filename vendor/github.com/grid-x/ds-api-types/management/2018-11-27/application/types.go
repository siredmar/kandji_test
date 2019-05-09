package v20181127

import (
	"github.com/grid-x/ds-api-types"
)

// Application represents the application as exported by this API version
type Application struct {
	Metadata types.Metadata `json:"metadata"`
	Name     string         `json:"name"`
}
