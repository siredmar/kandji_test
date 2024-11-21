package rule

import (
	"testing"

	v20190817Pod "github.com/grid-x/ds-api-types/management/2019-08-17/pod"
	deployments "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/pkg/api"
)

func TestDeploymentVolumesAllowed(t *testing.T) {
	testcases := []struct {
		rule      *DeploymentVolumesAllowed
		desc      string
		ctx       *context.Context
		res       interface{}
		wantPass  bool
		wantSkip  bool
		wantError bool
	}{
		{
			desc: "volumes allowed",
			rule: NewDeploymentVolumesAllowed([]string{"/tmp/goobaz"}),
			ctx:  nilCtx,
			res: &api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					App: "foo",
					Template: deployments.PodTemplate{
						Spec: v20190817Pod.PodConfig{
							Volumes: []v20190817Pod.Volume{
								{
									Name: "foovolume",
									VolumeSource: v20190817Pod.VolumeSource{
										HostPath: &v20190817Pod.HostPathVolumeSource{
											Path: "/tmp/foobar",
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
			desc: "volumes not allowed",
			rule: NewDeploymentVolumesAllowed([]string{"/tmp/goobaz"}),
			ctx:  nilCtx,
			res: &api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					App: "foo",
					Template: deployments.PodTemplate{
						Spec: v20190817Pod.PodConfig{
							Volumes: []v20190817Pod.Volume{
								{
									Name: "foovolume",
									VolumeSource: v20190817Pod.VolumeSource{
										HostPath: &v20190817Pod.HostPathVolumeSource{
											Path: "/tmp/foobar",
										},
									},
								},
								{
									Name: "goovolume",
									VolumeSource: v20190817Pod.VolumeSource{
										HostPath: &v20190817Pod.HostPathVolumeSource{
											Path: "/tmp/goobaz",
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
		{
			desc: "no volumes",
			rule: NewDeploymentVolumesAllowed([]string{"/tmp/goobaz"}),
			ctx:  nilCtx,
			res: &api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					App: "foo",
					Template: deployments.PodTemplate{
						Spec: v20190817Pod.PodConfig{
							Volumes: []v20190817Pod.Volume{},
						},
					},
				},
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			got, gotErr := tc.rule.Exec(tc.ctx, tc.res)
			got.SourceID = "test"
			got.Rule = tc.rule

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
