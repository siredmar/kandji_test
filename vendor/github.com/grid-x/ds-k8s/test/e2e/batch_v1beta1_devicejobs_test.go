package e2e

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	batchv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/batch/v1beta1"
)

var (
	batchv1beta1DeviceJobIgnores = IgnoreSuffixes(
		"UID",
		"SelfLink",
		"Time",
		"ResourceVersion",
		"Generation",
	)
)

func Test_BatchV1beta1_DeviceJob_Basic(t *testing.T) {
	client := GLOBAL.CRClient()

	ns, err := GLOBAL.NewTestNamespace()
	if err != nil {
		t.Fatalf("error while requesting test namespace: %+v", err)
	}

	testcases := []struct {
		wantErr bool
		job     *batchv1beta1.DeviceJob
	}{
		{
			wantErr: true,
			job: &batchv1beta1.DeviceJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
			},
		},
		{
			wantErr: true,
			job: &batchv1beta1.DeviceJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
				Spec:   batchv1beta1.DeviceJobSpec{},
				Status: batchv1beta1.DeviceJobStatus{},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			_, err := client.BatchV1beta1().DeviceJobs(ns).Create(tc.job)
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatalf("expected error but got nil")
			} else if tc.wantErr && err != nil {
				// that's fine :)
				return
			}

			got, err := client.BatchV1beta1().DeviceJobs(ns).Get(tc.job.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("error while getting device pod: %+v", err)
			}

			if !cmp.Equal(tc.job, got, batchv1beta1DeviceJobIgnores) {
				t.Errorf("unexpected device pod: %s", cmp.Diff(tc.job, got, batchv1beta1DeviceJobIgnores))
			}

			if err := client.BatchV1beta1().DeviceJobs(ns).Delete(tc.job.Name, nil); err != nil {
				t.Errorf("cannot delete device pod %s/%s: %+v", ns, tc.job.Name, err)
			}
		})
	}

}
