package e2e

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mainv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/maintenance/v1beta1"
)

var (
	mainv1beta1MaintenanceTaskIgnores = IgnoreSuffixes(
		"UID",
		"SelfLink",
		"Time",
		"ResourceVersion",
		"Generation",
	)
)

func Test_MainV1beta1_MaintenanceTask_Basic(t *testing.T) {
	client := GLOBAL.CRClient()

	ns, err := GLOBAL.NewTestNamespace()
	if err != nil {
		t.Fatalf("error while requesting test namespace: %+v", err)
	}

	testcases := []struct {
		wantErr bool
		task    *mainv1beta1.MaintenanceTask
	}{
		{
			wantErr: true,
			task: &mainv1beta1.MaintenanceTask{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
			},
		},
		{
			wantErr: true,
			task: &mainv1beta1.MaintenanceTask{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
				Spec:   mainv1beta1.MaintenanceTaskSpec{},
				Status: mainv1beta1.MaintenanceTaskStatus{},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			_, err := client.MaintenanceV1beta1().MaintenanceTasks(ns).Create(tc.task)
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatalf("expected error but got nil")
			} else if tc.wantErr && err != nil {
				// that's fine :)
				return
			}

			got, err := client.MaintenanceV1beta1().MaintenanceTasks(ns).Get(tc.task.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("error while getting device pod: %+v", err)
			}

			if !cmp.Equal(tc.task, got, mainv1beta1MaintenanceTaskIgnores) {
				t.Errorf("unexpected device pod: %s", cmp.Diff(tc.task, got, mainv1beta1MaintenanceTaskIgnores))
			}

			if err := client.MaintenanceV1beta1().MaintenanceTasks(ns).Delete(tc.task.Name, nil); err != nil {
				t.Errorf("cannot delete device pod %s/%s: %+v", ns, tc.task.Name, err)
			}
		})
	}

}
