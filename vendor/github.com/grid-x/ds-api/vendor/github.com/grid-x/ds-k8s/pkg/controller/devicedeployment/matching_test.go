package devicedeployment

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
)

const (
	GridXChannelLabel  = "gridx.de/channel"
	GridXStableChannel = "stable"
)

func createGBx(name string, labels map[string]string) corev1beta1.Device {
	return corev1beta1.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: corev1beta1.DeviceSpec{},
		Status: corev1beta1.DeviceStatus{
			LastHeartbeat: mkString(time.Now().Format(time.RFC3339)),
		},
	}
}

func createGBxD(name string, selector appsv1beta1.Selector) appsv1beta1.DeviceDeployment {
	return appsv1beta1.DeviceDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1beta1.DeviceDeploymentSpec{
			Selector: selector,
		},
		Status: appsv1beta1.DeviceDeploymentStatus{},
	}
}

func mkString(s string) *string {
	r := new(string)
	*r = s
	return r
}

func Test_IDMatches(t *testing.T) {
	testcases := []struct {
		gbx  corev1beta1.Device
		gbxd appsv1beta1.DeviceDeployment
		want bool
	}{
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByDeviceID: mkString("gridbox001"),
				}),
			want: true,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByDeviceID: mkString("gridbox002"),
				}),
			want: false,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByLabels: map[string]string{
						"foo": "bar",
					},
				}),
			want: false,
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("Test_IDMatches_%d", i), func(t *testing.T) {
			got := IDMatches(tc.gbx, tc.gbxd.Spec.Selector)
			if got != tc.want {
				t.Errorf("Expected %t, but got %t", tc.want, got)
			}
		})
	}
}
func Test_LabelsMatch(t *testing.T) {
	testcases := []struct {
		gbx  corev1beta1.Device
		gbxd appsv1beta1.DeviceDeployment
		want bool
	}{
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByDeviceID: mkString("gridbox001"),
				}),
			want: false,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByDeviceID: mkString("gridbox002"),
				}),
			want: false,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByLabels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
					},
				}),
			want: true,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByLabels: map[string]string{
						GridXChannelLabel: "alpha",
					},
				}),
			want: false,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByLabels: map[string]string{
						GridXChannelLabel: "alpha",
						"foo":             "bar",
					},
				}),
			want: false,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByDeviceID: mkString("gridbox001"),
					MatchByLabels: map[string]string{
						GridXChannelLabel: "alpha",
						"foo":             "bar",
					},
				}),
			want: false,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByDeviceID: mkString("gridbox001"),
					MatchByLabels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
						"foo":             "bar",
					},
				}),
			want: true,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByLabels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
						"foo":             "bar",
					},
				}),
			want: true,
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("Test_LabelsMatch_%d", i), func(t *testing.T) {
			got := LabelsMatch(tc.gbx, tc.gbxd.Spec.Selector)
			if got != tc.want {
				t.Errorf("Expected %t, but got %t", tc.want, got)
			}
		})
	}
}
func Test_Matches(t *testing.T) {
	testcases := []struct {
		gbx  corev1beta1.Device
		gbxd appsv1beta1.DeviceDeployment
		want bool
	}{
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByDeviceID: mkString("gridbox001"),
				}),
			want: true,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByDeviceID: mkString("gridbox002"),
				}),
			want: false,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByLabels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
					},
				}),
			want: true,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByLabels: map[string]string{
						GridXChannelLabel: "alpha",
					},
				}),
			want: false,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByLabels: map[string]string{
						GridXChannelLabel: "alpha",
						"foo":             "bar",
					},
				}),
			want: false,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByDeviceID: mkString("gridbox001"),
					MatchByLabels: map[string]string{
						GridXChannelLabel: "alpha",
						"foo":             "bar",
					},
				}),
			want: true,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByDeviceID: mkString("gridbox001"),
					MatchByLabels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
						"foo":             "bar",
					},
				}),
			want: true,
		},
		{
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxd: createGBxD(
				"d51bb63c-c20a-487e-be8f-26f6e704fbb3",
				appsv1beta1.Selector{
					MatchByLabels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
						"foo":             "bar",
					},
				}),
			want: true,
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("Test_Matches_%d", i), func(t *testing.T) {
			got := Matches(tc.gbx, tc.gbxd.Spec.Selector)
			if got != tc.want {
				t.Errorf("Expected %t, but got %t", tc.want, got)
			}
		})
	}
}

func Test_Precision(t *testing.T) {
	testcases := []struct {
		gbxd appsv1beta1.DeviceDeployment
		want int
	}{
		{
			gbxd: createGBxD(
				"58626771-b540-4921-9eec-65499a9b2e1a",
				appsv1beta1.Selector{
					MatchByLabels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
						"foo":             "bar",
					},
				},
			),
			want: 2,
		},
		{
			gbxd: createGBxD(
				"58626771-b540-4921-9eec-65499a9b2e1a",
				appsv1beta1.Selector{
					MatchByDeviceID: mkString("72561212-52d1-48be-8524-1f0490a4f1c0"),
				},
			),
			want: 0,
		},
		{
			gbxd: createGBxD(
				"58626771-b540-4921-9eec-65499a9b2e1a",
				appsv1beta1.Selector{
					MatchByDeviceID: mkString("72561212-52d1-48be-8524-1f0490a4f1c0"),
					MatchByLabels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
					},
				},
			),
			want: 1,
		},
		{
			gbxd: createGBxD(
				"58626771-b540-4921-9eec-65499a9b2e1a",
				appsv1beta1.Selector{
					MatchByLabels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
					},
				},
			),
			want: 1,
		},
		{
			gbxd: createGBxD(
				"58626771-b540-4921-9eec-65499a9b2e1a",
				appsv1beta1.Selector{
					MatchByLabels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
						"foo":             GridXStableChannel,
						"bar":             GridXStableChannel,
						"qaz":             GridXStableChannel,
					},
				},
			),
			want: 4,
		},
	}
	for i, tc := range testcases {
		t.Run(fmt.Sprintf("Test_Precision_%d", i), func(t *testing.T) {
			got := precision(tc.gbxd)
			if !cmp.Equal(tc.want, got) {
				t.Errorf("Expected %+v, but got %+v", tc.want, got)
			}
		})
	}
}

func Test_LastUpdatedAt(t *testing.T) {
	// Feeling brave and ignoring the error
	ts, _ := time.Parse(time.RFC3339, "2017-11-21T13:37:42")
	testcases := []struct {
		gbxd appsv1beta1.DeviceDeployment
		want time.Time
	}{
		{
			gbxd: appsv1beta1.DeviceDeployment{
				Status: appsv1beta1.DeviceDeploymentStatus{
					LastUpdatedAt: ts.Format(time.RFC3339),
				},
			},
			want: ts,
		},
	}
	for i, tc := range testcases {
		t.Run(fmt.Sprintf("Test_LastUpdatedAt_%d", i), func(t *testing.T) {
			got, err := lastUpdatedAt(tc.gbxd)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(tc.want, got) {
				t.Errorf("Expected %+v, but got %+v", tc.want, got)
			}
		})
	}
}

func Test_FindMatchingDeployments(t *testing.T) {

	now := time.Now().Format(time.RFC3339)
	yesterday := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)

	testcases := []struct {
		gbx   corev1beta1.Device
		gbxds []appsv1beta1.DeviceDeployment
		want  map[string]appsv1beta1.DeviceDeployment // indexed by app
	}{
		{
			// No deployments test
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxds: []appsv1beta1.DeviceDeployment{},
			want:  map[string]appsv1beta1.DeviceDeployment{},
		},
		{
			// No matching deployments test
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxds: []appsv1beta1.DeviceDeployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-alpha-channel",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: "alpha",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
			want: map[string]appsv1beta1.DeviceDeployment{},
		},
		{
			// No matching deployments test
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxds: []appsv1beta1.DeviceDeployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-alpha-channel",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: "alpha",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-alpha-channel-solaredge",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: "alpha",
								"vendor":          "solaredge",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
			want: map[string]appsv1beta1.DeviceDeployment{},
		},
		{
			// One matching deployments test
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxds: []appsv1beta1.DeviceDeployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-stable-channel",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
			want: map[string]appsv1beta1.DeviceDeployment{
				"monitoring": appsv1beta1.DeviceDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-stable-channel",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
		},
		{
			// "matchByID has higher prio" test
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxds: []appsv1beta1.DeviceDeployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-stable-channel",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-fix007-gbx001",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByDeviceID: mkString("gridbox001"),
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
			want: map[string]appsv1beta1.DeviceDeployment{
				"monitoring": appsv1beta1.DeviceDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-fix007-gbx001",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByDeviceID: mkString("gridbox001"),
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
		},
		{
			// two apps
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxds: []appsv1beta1.DeviceDeployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-stable-channel",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-exporter-stable",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "node-exporter",
						Selector: appsv1beta1.Selector{
							MatchByDeviceID: mkString("gridbox001"),
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
			want: map[string]appsv1beta1.DeviceDeployment{
				"monitoring": appsv1beta1.DeviceDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-stable-channel",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				"node-exporter": appsv1beta1.DeviceDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-exporter-stable",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "node-exporter",
						Selector: appsv1beta1.Selector{
							MatchByDeviceID: mkString("gridbox001"),
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
		},
		{
			// two apps
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxds: []appsv1beta1.DeviceDeployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-stable-channel",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: "foo",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-exporter-stable",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "node-exporter",
						Selector: appsv1beta1.Selector{
							MatchByDeviceID: mkString("gridbox001"),
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
			want: map[string]appsv1beta1.DeviceDeployment{
				"node-exporter": appsv1beta1.DeviceDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-exporter-stable",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "node-exporter",
						Selector: appsv1beta1.Selector{
							MatchByDeviceID: mkString("gridbox001"),
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
		},
		{
			// three apps
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"foo":             "bar",
			}),
			gbxds: []appsv1beta1.DeviceDeployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-stable-channel",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-experimental-channel",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: "experimental",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-solaredge",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
								"vendor":          "solaredge",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-exporter-stable",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "node-exporter",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-exporter-gbx001",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "node-exporter",
						Selector: appsv1beta1.Selector{
							MatchByDeviceID: mkString("gridbox001"),
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "envscan-stable",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "envscan",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "envscan-stable-sma",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "envscan",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
								"vendor":          "sma",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
			want: map[string]appsv1beta1.DeviceDeployment{
				"node-exporter": appsv1beta1.DeviceDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-exporter-gbx001",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "node-exporter",
						Selector: appsv1beta1.Selector{
							MatchByDeviceID: mkString("gridbox001"),
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				"monitoring": appsv1beta1.DeviceDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-stable-channel",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				"envscan": appsv1beta1.DeviceDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "envscan-stable",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "envscan",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
		},
		{
			// three apps with more precise deployment for envscan
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"vendor":          "sma",
			}),
			gbxds: []appsv1beta1.DeviceDeployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-stable-channel",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-experimental-channel",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: "experimental",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-solaredge",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
								"vendor":          "solaredge",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-exporter-stable",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "node-exporter",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-exporter-gbx001",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "node-exporter",
						Selector: appsv1beta1.Selector{
							MatchByDeviceID: mkString("gridbox001"),
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "envscan-stable",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "envscan",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "envscan-stable-sma",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "envscan",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
								"vendor":          "sma",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
			want: map[string]appsv1beta1.DeviceDeployment{
				"node-exporter": appsv1beta1.DeviceDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-exporter-gbx001",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "node-exporter",
						Selector: appsv1beta1.Selector{
							MatchByDeviceID: mkString("gridbox001"),
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				"monitoring": appsv1beta1.DeviceDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "monitoring-stable-channel",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "monitoring",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				"envscan": appsv1beta1.DeviceDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "envscan-stable-sma",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "envscan",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
								"vendor":          "sma",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
		},
		{
			// envscan app with equal precision
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"vendor":          "sma",
			}),
			gbxds: []appsv1beta1.DeviceDeployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "envscan-stable",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "envscan",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "envscan-stable-sma",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "envscan",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
								"vendor":          "sma",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: yesterday,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "envscan-stable-sma-fake-release",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "envscan",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
								"vendor":          "sma",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
			want: map[string]appsv1beta1.DeviceDeployment{
				"envscan": appsv1beta1.DeviceDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "envscan-stable-sma-fake-release",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "envscan",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
								"vendor":          "sma",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
		},
		{
			// envscan app with equal ID matching
			gbx: createGBx("gridbox001", map[string]string{
				GridXChannelLabel: GridXStableChannel,
				"vendor":          "sma",
			}),
			gbxds: []appsv1beta1.DeviceDeployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "envscan-stable",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "envscan",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
							MatchByDeviceID: mkString("gridbox001"),
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "envscan-stable-sma",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "envscan",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
								"vendor":          "sma",
							},
							MatchByDeviceID: mkString("gridbox001"),
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: yesterday,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "envscan-stable-sma-fake-release",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "envscan",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
								"vendor":          "sma",
							},
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
			want: map[string]appsv1beta1.DeviceDeployment{
				"envscan": appsv1beta1.DeviceDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "envscan-stable",
					},
					Spec: appsv1beta1.DeviceDeploymentSpec{
						App: "envscan",
						Selector: appsv1beta1.Selector{
							MatchByLabels: map[string]string{
								GridXChannelLabel: GridXStableChannel,
							},
							MatchByDeviceID: mkString("gridbox001"),
						},
					},
					Status: appsv1beta1.DeviceDeploymentStatus{
						LastUpdatedAt: now,
					},
				},
			},
		},
	}
	for i, tc := range testcases {
		t.Run(fmt.Sprintf("Test_FindMatchingDeployments_%d", i), func(t *testing.T) {
			got := findMatchingDeployments(tc.gbx, tc.gbxds)
			if len(got) != len(tc.want) {
				t.Fatalf("Unequal length. Wanted %d, but got %d items", len(tc.want), len(got))
			}
			for i, v := range tc.want {
				if !cmp.Equal(v, got[i]) {
					//t.Errorf("Expected %+v, but got %+v", v, got[i])
					t.Errorf("%s", cmp.Diff(v, got[i]))
				}
			}
		})
	}
}

func Test_FindMatchingContainers(t *testing.T) {
	testcases := []struct {
		gbx   corev1beta1.Device
		gbxcs []corev1beta1.DevicePod
		want  map[string][]corev1beta1.DevicePod
	}{
		{
			gbxcs: []corev1beta1.DevicePod{},
			want:  map[string][]corev1beta1.DevicePod{},
		},
		{
			want: map[string][]corev1beta1.DevicePod{},
		},
		{
			gbxcs: []corev1beta1.DevicePod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "23481a77-589a-4879-82c6-716bf46690d1",
						OwnerReferences: []metav1.OwnerReference{
							// Skipped for testing purposes
						},
						Annotations: map[string]string{
							appsv1beta1.AppNameAnnotation: "monitoring",
						},
					},
					Spec: corev1beta1.DevicePodSpec{
						DeviceID: "GridBox001",
						Config: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Image:   "gridx.de/monitoring:stable",
									Command: []string{"/usr/local/bin/monitoring"},
								},
							},
							Network: corev1beta1.NetworkSettingHost,
						},
					},
					Status: corev1beta1.DevicePodStatus{},
				},
			},
			want: map[string][]corev1beta1.DevicePod{},
		},
		{
			gbx: corev1beta1.Device{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gridbox002",
					Labels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
					},
				},
				Spec:   corev1beta1.DeviceSpec{},
				Status: corev1beta1.DeviceStatus{},
			},
			gbxcs: []corev1beta1.DevicePod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "23481a77-589a-4879-82c6-716bf46690d1",
						OwnerReferences: []metav1.OwnerReference{
							// Skipped for testing purposes
						},
						Annotations: map[string]string{
							appsv1beta1.AppNameAnnotation: "monitoring",
						},
					},
					Spec: corev1beta1.DevicePodSpec{
						DeviceID: "GridBox001",
						Config: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Image:   "gridx.de/monitoring:stable",
									Command: []string{"/usr/local/bin/monitoring"},
								},
							},
							Network: corev1beta1.NetworkSettingHost,
						},
					},
					Status: corev1beta1.DevicePodStatus{},
				},
			},
			want: map[string][]corev1beta1.DevicePod{},
		},
		{
			gbx: corev1beta1.Device{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gridbox002",
					Labels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
					},
				},
				Spec:   corev1beta1.DeviceSpec{},
				Status: corev1beta1.DeviceStatus{},
			},
			gbxcs: nil,
			want:  map[string][]corev1beta1.DevicePod{},
		},
		{
			gbx: corev1beta1.Device{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gridbox001",
					Labels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
					},
				},
				Spec:   corev1beta1.DeviceSpec{},
				Status: corev1beta1.DeviceStatus{},
			},
			gbxcs: []corev1beta1.DevicePod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "23481a77-589a-4879-82c6-716bf46690d1",
						OwnerReferences: []metav1.OwnerReference{
							// Skipped for testing purposes
						},
						Annotations: map[string]string{
							appsv1beta1.AppNameAnnotation: "monitoring",
						},
					},
					Spec: corev1beta1.DevicePodSpec{
						DeviceID: "gridbox001",
						Config: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Image:   "gridx.de/monitoring:stable",
									Command: []string{"/usr/local/bin/monitoring"},
								},
							},
							Network: corev1beta1.NetworkSettingHost,
						},
					},
					Status: corev1beta1.DevicePodStatus{},
				},
			},
			want: map[string][]corev1beta1.DevicePod{
				"monitoring": []corev1beta1.DevicePod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "23481a77-589a-4879-82c6-716bf46690d1",
							OwnerReferences: []metav1.OwnerReference{},
							Annotations: map[string]string{
								appsv1beta1.AppNameAnnotation: "monitoring",
							},
						},
						Spec: corev1beta1.DevicePodSpec{
							DeviceID: "gridbox001",
							Config: corev1beta1.PodConfig{
								Containers: []corev1beta1.Container{
									{
										Image:   "gridx.de/monitoring:stable",
										Command: []string{"/usr/local/bin/monitoring"},
									},
								},
								Network: corev1beta1.NetworkSettingHost,
							},
						},
					},
				},
			},
		},
		{
			gbx: corev1beta1.Device{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gridbox001",
					Labels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
					},
				},
				Spec:   corev1beta1.DeviceSpec{},
				Status: corev1beta1.DeviceStatus{},
			},
			gbxcs: []corev1beta1.DevicePod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "23481a77-589a-4879-82c6-716bf46690d1",
						OwnerReferences: []metav1.OwnerReference{
							// Skipped for testing purposes
						},
						Annotations: map[string]string{
							appsv1beta1.AppNameAnnotation: "monitoring",
						},
					},
					Spec: corev1beta1.DevicePodSpec{
						DeviceID: "gridbox001",
						Config: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Image:   "gridx.de/monitoring:stable",
									Command: []string{"/usr/local/bin/monitoring"},
								},
							},
							Network: corev1beta1.NetworkSettingHost,
						},
					},
					Status: corev1beta1.DevicePodStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "d1b7cab2-92cc-4e9d-a041-edb311c729a3",
						OwnerReferences: []metav1.OwnerReference{
							// Skipped for testing purposes
						},
						Annotations: map[string]string{
							appsv1beta1.AppNameAnnotation: "monitoring",
						},
					},
					Spec: corev1beta1.DevicePodSpec{
						DeviceID: "gridbox002",
						Config: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Image:   "gridx.de/monitoring:stable",
									Command: []string{"/usr/local/bin/monitoring"},
								},
							},
							Network: corev1beta1.NetworkSettingHost,
						},
					},
					Status: corev1beta1.DevicePodStatus{},
				},
			},
			want: map[string][]corev1beta1.DevicePod{
				"monitoring": []corev1beta1.DevicePod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "23481a77-589a-4879-82c6-716bf46690d1",
							OwnerReferences: []metav1.OwnerReference{},
							Annotations: map[string]string{
								appsv1beta1.AppNameAnnotation: "monitoring",
							},
						},
						Spec: corev1beta1.DevicePodSpec{
							DeviceID: "gridbox001",
							Config: corev1beta1.PodConfig{
								Containers: []corev1beta1.Container{
									{
										Image:   "gridx.de/monitoring:stable",
										Command: []string{"/usr/local/bin/monitoring"},
									},
								},
								Network: corev1beta1.NetworkSettingHost,
							},
						},
					},
				},
			},
		},
		{
			gbx: corev1beta1.Device{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gridbox001",
					Labels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
					},
				},
				Spec:   corev1beta1.DeviceSpec{},
				Status: corev1beta1.DeviceStatus{},
			},
			gbxcs: []corev1beta1.DevicePod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "23481a77-589a-4879-82c6-716bf46690d1",
						OwnerReferences: []metav1.OwnerReference{
							// Skipped for testing purposes
						},
						Annotations: map[string]string{
							appsv1beta1.AppNameAnnotation: "monitoring",
						},
					},
					Spec: corev1beta1.DevicePodSpec{
						DeviceID: "gridbox001",
						Config: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Image:   "gridx.de/monitoring:stable",
									Command: []string{"/usr/local/bin/monitoring"},
								},
							},
							Network: corev1beta1.NetworkSettingHost,
						},
					},
					Status: corev1beta1.DevicePodStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "d1b7cab2-92cc-4e9d-a041-edb311c729a3",
						OwnerReferences: []metav1.OwnerReference{
							// Skipped for testing purposes
						},
						Annotations: map[string]string{
							appsv1beta1.AppNameAnnotation: "monitoring",
						},
					},
					Spec: corev1beta1.DevicePodSpec{
						DeviceID: "gridbox002",
						Config: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Image:   "gridx.de/monitoring:stable",
									Command: []string{"/usr/local/bin/monitoring"},
								},
							},
							Network: corev1beta1.NetworkSettingHost,
						},
					},
					Status: corev1beta1.DevicePodStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "b5d9a5de-6a32-4dcc-ad1a-1a7655ae216e",
						Annotations: map[string]string{
							appsv1beta1.AppNameAnnotation: "envscan",
						},
						OwnerReferences: []metav1.OwnerReference{
							// Skipped for testing purposes
						},
					},
					Spec: corev1beta1.DevicePodSpec{
						DeviceID: "gridbox001",
						Config: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Image:   "gridx.de/envscan:stable",
									Command: []string{"/usr/local/bin/envscan"},
								},
							},
							Network: corev1beta1.NetworkSettingHost,
						},
					},
					Status: corev1beta1.DevicePodStatus{},
				},
			},
			want: map[string][]corev1beta1.DevicePod{
				"monitoring": []corev1beta1.DevicePod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "23481a77-589a-4879-82c6-716bf46690d1",
							Annotations: map[string]string{
								appsv1beta1.AppNameAnnotation: "monitoring",
							},
							OwnerReferences: []metav1.OwnerReference{},
						},
						Spec: corev1beta1.DevicePodSpec{
							DeviceID: "gridbox001",
							Config: corev1beta1.PodConfig{
								Containers: []corev1beta1.Container{
									{
										Image:   "gridx.de/monitoring:stable",
										Command: []string{"/usr/local/bin/monitoring"},
									},
								},
								Network: corev1beta1.NetworkSettingHost,
							},
						},
					},
				},
				"envscan": []corev1beta1.DevicePod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "b5d9a5de-6a32-4dcc-ad1a-1a7655ae216e",
							Annotations: map[string]string{
								appsv1beta1.AppNameAnnotation: "envscan",
							},
							OwnerReferences: []metav1.OwnerReference{
								// Skipped for testing purposes
							},
						},
						Spec: corev1beta1.DevicePodSpec{
							DeviceID: "gridbox001",
							Config: corev1beta1.PodConfig{
								Containers: []corev1beta1.Container{
									{
										Image:   "gridx.de/envscan:stable",
										Command: []string{"/usr/local/bin/envscan"},
									},
								},
								Network: corev1beta1.NetworkSettingHost,
							},
						},
						Status: corev1beta1.DevicePodStatus{},
					},
				},
			},
		},
		{
			gbx: corev1beta1.Device{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gridbox002",
					Labels: map[string]string{
						GridXChannelLabel: GridXStableChannel,
					},
				},
				Spec:   corev1beta1.DeviceSpec{},
				Status: corev1beta1.DeviceStatus{},
			},
			gbxcs: []corev1beta1.DevicePod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "23481a77-589a-4879-82c6-716bf46690d1",
						Annotations: map[string]string{
							appsv1beta1.AppNameAnnotation: "monitoring",
						},
						OwnerReferences: []metav1.OwnerReference{
							// Skipped for testing purposes
						},
					},
					Spec: corev1beta1.DevicePodSpec{
						DeviceID: "gridbox001",
						Config: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Image:   "gridx.de/monitoring:stable",
									Command: []string{"/usr/local/bin/monitoring"},
								},
							},
							Network: corev1beta1.NetworkSettingHost,
						},
					},
					Status: corev1beta1.DevicePodStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "d1b7cab2-92cc-4e9d-a041-edb311c729a3",
						Annotations: map[string]string{
							appsv1beta1.AppNameAnnotation: "monitoring",
						},
						OwnerReferences: []metav1.OwnerReference{
							// Skipped for testing purposes
						},
					},
					Spec: corev1beta1.DevicePodSpec{
						DeviceID: "gridbox002",
						Config: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Image:   "gridx.de/monitoring:stable",
									Command: []string{"/usr/local/bin/monitoring"},
								},
							},
							Network: corev1beta1.NetworkSettingHost,
						},
					},
					Status: corev1beta1.DevicePodStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "b5d9a5de-6a32-4dcc-ad1a-1a7655ae216e",
						Annotations: map[string]string{
							appsv1beta1.AppNameAnnotation: "envscan",
						},
						OwnerReferences: []metav1.OwnerReference{
							// Skipped for testing purposes
						},
					},
					Spec: corev1beta1.DevicePodSpec{
						DeviceID: "gridbox001",
						Config: corev1beta1.PodConfig{
							Containers: []corev1beta1.Container{
								{
									Image:   "gridx.de/envscan:stable",
									Command: []string{"/usr/local/bin/envscan"},
								},
							},
							Network: corev1beta1.NetworkSettingHost,
						},
					},
					Status: corev1beta1.DevicePodStatus{},
				},
			},
			want: map[string][]corev1beta1.DevicePod{
				"monitoring": []corev1beta1.DevicePod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "d1b7cab2-92cc-4e9d-a041-edb311c729a3",
							Annotations: map[string]string{
								appsv1beta1.AppNameAnnotation: "monitoring",
							},
							OwnerReferences: []metav1.OwnerReference{},
						},
						Spec: corev1beta1.DevicePodSpec{
							DeviceID: "gridbox002",
							Config: corev1beta1.PodConfig{
								Containers: []corev1beta1.Container{
									{
										Image:   "gridx.de/monitoring:stable",
										Command: []string{"/usr/local/bin/monitoring"},
									},
								},
								Network: corev1beta1.NetworkSettingHost,
							},
						},
					},
				},
			},
		},
	}

	for k, tc := range testcases {
		t.Run(fmt.Sprintf("Test_FindMatchingContainers_%d", k), func(t *testing.T) {
			got := findMatchingContainers(tc.gbx, tc.gbxcs)
			if !cmp.Equal(tc.want, got) {
				//t.Errorf("Expected %+v, but got %+v", v, got[i])
				t.Errorf("%s", cmp.Diff(tc.want, got))
			}
		})
	}
}
