package e2e

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
)

var (
	appsv1beta1DeviceApplicationIgnores = IgnoreSuffixes(
		"UID",
		"SelfLink",
		"Time",
		"ResourceVersion",
		"Generation",
	)
)

func Test_Appsv1beta1_DeviceApplication_Basic(t *testing.T) {
	client := GLOBAL.CRClient()

	ns, err := GLOBAL.NewTestNamespace()
	if err != nil {
		t.Fatalf("error while requesting test namespace: %+v", err)
	}

	testcases := []struct {
		wantErr bool
		dev     *appsv1beta1.DeviceApplication
	}{
		{
			dev: &appsv1beta1.DeviceApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
			},
		},
		{
			dev: &appsv1beta1.DeviceApplication{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
				Spec:   appsv1beta1.DeviceApplicationSpec{},
				Status: appsv1beta1.DeviceApplicationStatus{},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			_, err := client.AppsV1beta1().DeviceApplications(ns).Create(tc.dev)
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatalf("expected error but got nil")
			} else if tc.wantErr && err != nil {
				// that's fine :)
				return
			}

			got, err := client.AppsV1beta1().DeviceApplications(ns).Get(tc.dev.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("error while getting device pod: %+v", err)
			}

			if !cmp.Equal(tc.dev, got, appsv1beta1DeviceApplicationIgnores) {
				t.Errorf("unexpected device pod: %s", cmp.Diff(tc.dev, got, appsv1beta1DeviceApplicationIgnores))
			}

			if err := client.AppsV1beta1().DeviceApplications(ns).Delete(tc.dev.Name, nil); err != nil {
				t.Errorf("cannot delete device pod %s/%s: %+v", ns, tc.dev.Name, err)
			}
		})
	}

}
