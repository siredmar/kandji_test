package e2e

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	configv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/config/v1beta1"
)

var (
	configv1beta1DockerConfigIgnores = IgnoreSuffixes(
		"UID",
		"SelfLink",
		"Time",
		"ResourceVersion",
		"Generation",
	)
)

func Test_Configv1beta1_DockerConfig_Basic(t *testing.T) {
	client := GLOBAL.CRClient()

	ns, err := GLOBAL.NewTestNamespace()
	if err != nil {
		t.Fatalf("error while requesting test namespace: %+v", err)
	}

	testcases := []struct {
		wantErr bool
		c       *configv1beta1.DockerConfig
	}{
		{
			wantErr: true,
			c: &configv1beta1.DockerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
			},
		},
		{
			wantErr: true,
			c: &configv1beta1.DockerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
				Spec:   configv1beta1.DockerConfigSpec{},
				Status: configv1beta1.DockerConfigStatus{},
			},
		},
		{
			c: &configv1beta1.DockerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foobar",
					Namespace: ns,
				},
				Spec: configv1beta1.DockerConfigSpec{
					Selector: appsv1beta1.Selector{
						MatchByLabels: map[string]string{
							"foo": "bar",
						},
					},
					Registry: "foo",
					Credentials: configv1beta1.DockerConfigCredentails{
						AWS: &configv1beta1.AWSCredentialProvider{
							AccessKey:       "foo",
							SecretAccessKey: "foo",
							Region:          "foo",
						},
					},
				},
				Status: configv1beta1.DockerConfigStatus{},
			},
		},
	}

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			_, err := client.ConfigV1beta1().DockerConfigs(ns).Create(tc.c)
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %+v", err)
			} else if tc.wantErr && err == nil {
				t.Fatalf("expected error but got nil")
			} else if tc.wantErr && err != nil {
				// that's fine :)
				return
			}

			got, err := client.ConfigV1beta1().DockerConfigs(ns).Get(tc.c.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("error while getting device pod: %+v", err)
			}

			if !cmp.Equal(tc.c, got, configv1beta1DockerConfigIgnores) {
				t.Errorf("unexpected device pod: %s", cmp.Diff(tc.c, got, configv1beta1DockerConfigIgnores))
			}

			if err := client.ConfigV1beta1().DockerConfigs(ns).Delete(tc.c.Name, nil); err != nil {
				t.Errorf("cannot delete device pod %s/%s: %+v", ns, tc.c.Name, err)
			}
		})
	}

}
