package e2e

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
)

var (
	corev1beta1DevicePodIgnores = IgnoreSuffixes(
		"UID",
		"SelfLink",
		"Time",
		"ResourceVersion",
		"Generation",
	)
)

func Test_CoreV1beta1_DevicePod_Basic(t *testing.T) {
	client := GLOBAL.CRClient()

	ns, err := GLOBAL.NewTestNamespace()
	if err != nil {
		t.Fatalf("error while requesting test namespace: %+v", err)
	}

	testcases := []struct {
		wantErr bool
		dev     *corev1beta1.DevicePod
	}{
		{
			wantErr: true,
			dev: &corev1beta1.DevicePod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
			},
		},
		{
			wantErr: true,
			dev: &corev1beta1.DevicePod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
				Spec:   corev1beta1.DevicePodSpec{},
				Status: corev1beta1.DevicePodStatus{},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			_, err := client.CoreV1beta1().DevicePods(ns).Create(tc.dev)
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatalf("expected error but got nil")
			} else if tc.wantErr && err != nil {
				// that's fine :)
				return
			}

			got, err := client.CoreV1beta1().DevicePods(ns).Get(tc.dev.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("error while getting device pod: %+v", err)
			}

			if !cmp.Equal(tc.dev, got, corev1beta1DevicePodIgnores) {
				t.Errorf("unexpected device pod: %s", cmp.Diff(tc.dev, got, corev1beta1DevicePodIgnores))
			}

			if err := client.CoreV1beta1().DevicePods(ns).Delete(tc.dev.Name, nil); err != nil {
				t.Errorf("cannot delete device pod %s/%s: %+v", ns, tc.dev.Name, err)
			}
		})
	}

}
