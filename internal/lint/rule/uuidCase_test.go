package rule

import (
	"fmt"
	"testing"

	types "github.com/grid-x/ds-api-types"
	maintenance "github.com/grid-x/ds-api-types/management/2019-11-04/maintenance"
	deployments "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"
	"github.com/grid-x/gxctl/pkg/api"
)

func TestUUIDs(t *testing.T) {
	r := UUIDCase{}

	testcases := []struct {
		desc      string
		res       interface{}
		wantPass  bool
		wantSkip  bool
		wantError bool
	}{
		// DEPLOYMENT
		{
			desc: "Deployment: both metadata.ID and spec.selector.matchByDeviceID have valid UUIDs",
			res: &api.Deployment{
				Metadata: types.Metadata{
					ID: "d562e404-8684-409d-94c1-9a8d0f52b7a7",
				},
				Spec: deployments.DeviceDeploymentSpec{
					Selector: deployments.Selector{
						MatchByDeviceID: func() *string {
							s := "d562e404-8684-409d-94c1-9a8d0f52b7a7"
							return &s
						}(),
					},
				},
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "Deployment: metadata.ID valid but spec.selector.matchByDeviceID has invalid UUID",
			res: &api.Deployment{
				Metadata: types.Metadata{
					ID: "d562e404-8684-409d-94c1-9a8d0f52b7a7",
				},
				Spec: deployments.DeviceDeploymentSpec{
					Selector: deployments.Selector{
						MatchByDeviceID: func() *string {
							s := "Aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
							return &s
						}(),
					},
				},
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "Deployment: metadata.ID invalid but spec.selector.matchByDeviceID has valid UUID",
			res: &api.Deployment{
				Metadata: types.Metadata{
					ID: "DDDDDDDD-8684-409d-94c1-9a8d0f52b7a7",
				},
				Spec: deployments.DeviceDeploymentSpec{
					Selector: deployments.Selector{
						MatchByDeviceID: func() *string {
							s := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
							return &s
						}(),
					},
				},
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
		// DEVICE
		{
			desc: "Device: valid metadata.ID",
			res: &api.Device{
				Metadata: types.Metadata{
					ID: "d562e404-8684-409d-94c1-9a8d0f52b7a7",
				},
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "Device: invalid metadata.ID",
			res: &api.Device{
				Metadata: types.Metadata{
					ID: "DDDDDDDD-8684-409d-94c1-9a8d0f52b7a7",
				},
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
		// MAINTENANCETASK
		{
			desc: "MaintenanceTask: all UUIDs valid",
			res: &api.MaintenanceTask{
				Metadata: types.Metadata{
					ID: "d562e404-8684-409d-94c1-9a8d0f52b7a7",
				},
				Spec: maintenance.MaintenanceTaskSpec{
					DeviceID: "d562e404-8684-409d-94c1-9a8d0f52b7a7",
				},
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "MaintenanceTask: metadata.ID invalid only",
			res: &api.MaintenanceTask{
				Metadata: types.Metadata{
					ID: "DDDDDDDD-8684-409d-94c1-9a8d0f52b7a7",
				},
				Spec: maintenance.MaintenanceTaskSpec{
					DeviceID: "d562e404-8684-409d-94c1-9a8d0f52b7a7",
				},
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "MaintenanceTask: spec.deviceID invalid only",
			res: &api.MaintenanceTask{
				Metadata: types.Metadata{
					ID: "d562e404-8684-409d-94c1-9a8d0f52b7a7",
				},
				Spec: maintenance.MaintenanceTaskSpec{
					DeviceID: "DDDDDDDD-8684-409d-94c1-9a8d0f52b7a7",
				},
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "MaintenanceTask: all UUIDs invalid",
			res: &api.MaintenanceTask{
				Metadata: types.Metadata{
					ID: "DDDDDDDD-8684-409d-94c1-9a8d0f52b7a7",
				},
				Spec: maintenance.MaintenanceTaskSpec{
					DeviceID: "DDDDDDDD-8684-409d-94c1-9a8d0f52b7a7",
				},
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%s", tc.desc), func(t *testing.T) {
			got, gotErr := r.Exec(nil, tc.res)
			got.SourceID = "test"
			got.Rule = &r

			didErr := false
			if tc.wantError && gotErr == nil || !tc.wantError && gotErr != nil {
				t.Errorf("wanted error=%v, got error=%v", tc.wantError, gotErr)
				didErr = true
			}
			if tc.wantPass != got.Pass {
				t.Errorf("wanted pass=%v, got pass=%v", tc.wantPass, got.Pass)
				didErr = true
			}
			if tc.wantSkip != got.Skip {
				t.Errorf("wanted skip=%v, got skip=%v", tc.wantSkip, got.Skip)
				didErr = true
			}
			if didErr {
				t.Errorf("have: %v", got.Have)
			}
		})
	}
}
