package rule

import (
	"fmt"
	"testing"

	types "github.com/grid-x/ds-api-types"
	v20190817Pod "github.com/grid-x/ds-api-types/management/2019-08-17/pod"
	deployments "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"

	deviceConfigMapApi "github.com/grid-x/ds-api-types/management/2021-03-10/deviceconfigmaps"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/state"
	"github.com/grid-x/gxctl/pkg/api"
)

func TestDeploymentDeviceConfigMapItemsExist(t *testing.T) {
	r := DeploymentDeviceConfigMapItemsExist{}

	testcases := []struct {
		desc      string
		ctx       *context.Context
		res       interface{}
		wantPass  bool
		wantSkip  bool
		wantError bool
	}{
		{
			desc: "dcm item exists in data",
			ctx: &context.Context{
				Current: state.State{
					DeviceConfigMaps: []api.DeviceConfigMap{
						{
							Metadata: types.Metadata{ID: "foo-dcm"},
							Spec: deviceConfigMapApi.DeviceConfigMapSpec{
								Immutable: new(bool),
								Data: map[string]string{
									"foo-key": "blub",
								},
								BinaryData: map[string][]byte{},
							},
							Status: deviceConfigMapApi.DeviceConfigMapStatus{},
						},
					},
				},
				Desired: nilState,
			},
			res: &api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					Template: deployments.PodTemplate{
						Spec: v20190817Pod.PodConfig{
							Volumes: []v20190817Pod.Volume{
								{
									VolumeSource: v20190817Pod.VolumeSource{
										ConfigMap: &v20190817Pod.ConfigMapVolumeSource{
											Name: "foo-dcm",
											Items: []v20190817Pod.KeyToPath{
												{
													Key:  "foo-key",
													Path: "bar",
													Mode: new(int32),
												},
											},
											DefaultMode: new(int32),
										},
									},
								},
							},
						},
					},
				},
			},
			wantPass:  true,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "dcm item exists in binary data",
			ctx: &context.Context{
				Current: state.State{
					DeviceConfigMaps: []api.DeviceConfigMap{
						{
							Metadata: types.Metadata{ID: "foo-dcm"},
							Spec: deviceConfigMapApi.DeviceConfigMapSpec{
								Immutable: new(bool),
								Data:      map[string]string{},
								BinaryData: map[string][]byte{
									"foo-key": []byte("blub"),
								},
							},
							Status: deviceConfigMapApi.DeviceConfigMapStatus{},
						},
					},
				},
				Desired: nilState,
			},
			res: &api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					Template: deployments.PodTemplate{
						Spec: v20190817Pod.PodConfig{
							Volumes: []v20190817Pod.Volume{
								{
									VolumeSource: v20190817Pod.VolumeSource{
										ConfigMap: &v20190817Pod.ConfigMapVolumeSource{
											Name: "foo-dcm",
											Items: []v20190817Pod.KeyToPath{
												{
													Key:  "foo-key",
													Path: "bar",
													Mode: new(int32),
												},
											},
											DefaultMode: new(int32),
										},
									},
								},
							},
						},
					},
				},
			},
			wantPass:  true,
			wantSkip:  false,
			wantError: false,
		},
		{
			desc: "dcm item dose not exist",
			ctx: &context.Context{
				Current: state.State{
					DeviceConfigMaps: []api.DeviceConfigMap{
						{
							Metadata: types.Metadata{ID: "foo-dcm"},
							Spec: deviceConfigMapApi.DeviceConfigMapSpec{
								Immutable: new(bool),
								Data: map[string]string{
									"goo-key": "blub",
								},
								BinaryData: map[string][]byte{},
							},
							Status: deviceConfigMapApi.DeviceConfigMapStatus{},
						},
					},
				},
				Desired: nilState,
			},
			res: &api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					Template: deployments.PodTemplate{
						Spec: v20190817Pod.PodConfig{
							Volumes: []v20190817Pod.Volume{
								{
									VolumeSource: v20190817Pod.VolumeSource{
										ConfigMap: &v20190817Pod.ConfigMapVolumeSource{
											Name: "foo-dcm",
											Items: []v20190817Pod.KeyToPath{
												{
													Key:  "foo-key",
													Path: "bar",
													Mode: new(int32),
												},
											},
											DefaultMode: new(int32),
										},
									},
								},
							},
						},
					},
				},
			},
			wantPass:  false,
			wantSkip:  false,
			wantError: false,
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%s", tc.desc), func(t *testing.T) {
			got, gotErr := r.Exec(tc.ctx, tc.res)

			didErr := false
			if tc.wantError && gotErr == nil || !tc.wantError && gotErr != nil {
				t.Errorf("wanted error=%v, got error=%v", tc.wantError, gotErr)
				didErr = true
			}

			got.SourceID = "test"
			got.Rule = &r

			if tc.wantPass != got.Pass {
				t.Errorf("wanted pass=%v, got pass=%v", tc.wantPass, got.Pass)
				didErr = true
			}
			if tc.wantSkip != got.Skip {
				t.Errorf("wanted skip=%v, got skip=%v", tc.wantSkip, got.Skip)
				didErr = true
			}
			if didErr {
				t.Errorf("want: %v", got.Want)
				t.Errorf("have: %v", got.Have)
			}
		})
	}
}
