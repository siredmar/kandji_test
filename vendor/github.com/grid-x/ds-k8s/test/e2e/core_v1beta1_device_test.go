package e2e

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
)

var (
	corev1beta1DeviceIgnores = IgnoreSuffixes(
		"UID",
		"SelfLink",
		"Time",
		"ResourceVersion",
		"Generation",
	)
)

func Test_CoreV1beta1_Device_Basic(t *testing.T) {
	client := GLOBAL.CRClient()

	ns, err := GLOBAL.NewTestNamespace()
	if err != nil {
		t.Fatalf("error while requesting test namespace: %+v", err)
	}

	testcases := []struct {
		wantErr bool
		dev     *corev1beta1.Device
	}{
		{
			wantErr: true,
			dev:     &corev1beta1.Device{},
		},
		{
			dev: &corev1beta1.Device{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
			},
		},
		{
			dev: &corev1beta1.Device{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
				Spec:   corev1beta1.DeviceSpec{},
				Status: corev1beta1.DeviceStatus{},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			_, err := client.CoreV1beta1().Devices(ns).Create(tc.dev)
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatalf("expected error but got nil")
			} else if tc.wantErr && err != nil {
				// that's fine :)
				return
			}

			got, err := client.CoreV1beta1().Devices(ns).Get(tc.dev.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("error while getting device: %+v", err)
			}

			if !cmp.Equal(tc.dev, got, corev1beta1DeviceIgnores) {
				t.Errorf("unexpected device: %s", cmp.Diff(tc.dev, got, corev1beta1DeviceIgnores))
			}

			if err := client.CoreV1beta1().Devices(ns).Delete(tc.dev.Name, nil); err != nil {
				t.Errorf("cannot delete device %s/%s: %+v", ns, tc.dev.Name, err)
			}
		})
	}

}
