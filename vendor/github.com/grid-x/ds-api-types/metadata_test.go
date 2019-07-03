package types

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_metadataFiltered(t *testing.T) {
	testcases := []struct {
		input      metav1.ObjectMeta
		unfiltered bool
		want       Metadata
	}{
		{
			input: metav1.ObjectMeta{
				Name:      "e8cd400e-c5c7-4cb6-947b-50ae1ba951bc",
				Namespace: "account-ff8a5861-824b-4246-9457-94f04d665d7b",
				Annotations: map[string]string{
					"gridx.ai/app": "nginx",
				},
			},
			unfiltered: false,
			want: Metadata{
				ID: "e8cd400e-c5c7-4cb6-947b-50ae1ba951bc",
				Annotations: map[string]string{
					"gridx.ai/app": "nginx",
				},
			},
		},
		{
			input: metav1.ObjectMeta{
				Name:      "e8cd400e-c5c7-4cb6-947b-50ae1ba951bc",
				Namespace: "account-ff8a5861-824b-4246-9457-94f04d665d7b",
				Annotations: map[string]string{
					"gridx.ai/app": "nginx",
					"kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"core.gridx.ai/v1beta1\"}",
				},
			},
			unfiltered: false,
			want: Metadata{
				ID: "e8cd400e-c5c7-4cb6-947b-50ae1ba951bc",
				Annotations: map[string]string{
					"gridx.ai/app": "nginx",
				},
			},
		},
		{
			input: metav1.ObjectMeta{
				Name:      "e8cd400e-c5c7-4cb6-947b-50ae1ba951bc",
				Namespace: "account-ff8a5861-824b-4246-9457-94f04d665d7b",
				Annotations: map[string]string{
					"core.gridx.ai/app": "nginx",
					"kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"core.gridx.ai/v1beta1\"}",
				},
			},
			unfiltered: true,
			want: Metadata{
				ID: "e8cd400e-c5c7-4cb6-947b-50ae1ba951bc",
				Annotations: map[string]string{
					"core.gridx.ai/app": "nginx",
					"kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"core.gridx.ai/v1beta1\"}",
				},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := ConvertFromK8sMetadata(tc.input, tc.unfiltered)

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected metadata %s", cmp.Diff(tc.want, got))
			}
		})
	}
}
