package rule

import (
	"fmt"
	"testing"

	v20190817Pod "github.com/grid-x/ds-api-types/management/2019-08-17/pod"
	deployments "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/pkg/api"
)

func TestDeploymentVolumeMountsAllowed(t *testing.T) {
	testcases := []struct {
		rule      *DeploymentVolumeMountsAllowed
		desc      string
		ctx       *context.Context
		res       interface{}
		wantPass  bool
		wantSkip  bool
		wantError bool
	}{
		{
			desc: "allowed volumeMounts",
			rule: NewDeploymentVolumeMountsAllowed([]string{"goobaz"}),
			ctx:  nilCtx,
			res: &api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					App: "foo",
					Template: deployments.PodTemplate{
						Spec: v20190817Pod.PodConfig{
							Containers: []v20190817Pod.Container{
								{
									VolumeMounts: []v20190817Pod.VolumeMount{
										{
											Name:      "foobar",
											MountPath: "/etc/foobar",
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
			desc: "disallowed volumeMounts",
			rule: NewDeploymentVolumeMountsAllowed([]string{"/etc/goobaz", "/dev/blub"}),
			ctx:  nilCtx,
			res: &api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					App: "foo",
					Template: deployments.PodTemplate{
						Spec: v20190817Pod.PodConfig{
							Containers: []v20190817Pod.Container{
								{
									VolumeMounts: []v20190817Pod.VolumeMount{
										{
											Name:      "blub",
											MountPath: "/dev/blub",
										},
										{
											Name:      "foobar",
											MountPath: "/etc/foobar",
										},
									},
								},
								{
									VolumeMounts: []v20190817Pod.VolumeMount{
										{
											Name:      "foobar",
											MountPath: "/etc/foobar",
										},
										{
											Name:      "goobaz",
											MountPath: "/etc/goobaz",
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
			desc: "no containers",
			rule: NewDeploymentVolumeMountsAllowed([]string{"goobaz"}),
			ctx:  nilCtx,
			res: &api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					App: "app-foo",
					Template: deployments.PodTemplate{
						Spec: v20190817Pod.PodConfig{
							Containers: []v20190817Pod.Container{},
						},
					},
				},
			},
			wantPass:  true,
			wantSkip:  true,
			wantError: false,
		},
		{
			desc: "no volumeMounts",
			rule: NewDeploymentVolumeMountsAllowed([]string{"/etc/goobaz"}),
			ctx:  nilCtx,
			res: &api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					App: "foo",
					Template: deployments.PodTemplate{
						Spec: v20190817Pod.PodConfig{
							Containers: []v20190817Pod.Container{
								{
									VolumeMounts: []v20190817Pod.VolumeMount{},
								},
							},
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
		t.Run(fmt.Sprintf("%s", tc.desc), func(t *testing.T) {
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
