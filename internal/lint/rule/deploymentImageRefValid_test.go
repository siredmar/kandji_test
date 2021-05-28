package rule

import (
	"fmt"
	"strings"
	"testing"

	v20190817Pod "github.com/grid-x/ds-api-types/management/2019-08-17/pod"
	deployments "github.com/grid-x/ds-api-types/management/2020-08-29/deployments"

	"github.com/grid-x/gxctl/pkg/api"
)

func TestDeploymentImageRefValid(t *testing.T) {
	testcases := []struct {
		ref      string
		wantPass bool
	}{
		{
			ref:      "",
			wantPass: false,
		},
		{
			ref:      "short",
			wantPass: true,
		},
		{
			ref:      "simple/name",
			wantPass: true,
		},
		{
			ref:      "library/ubuntu",
			wantPass: true,
		},
		{
			ref:      "docker/stevvooe/app",
			wantPass: true,
		},
		{
			ref:      "aa/aa/aa/aa/aa/aa/aa/aa/aa/bb/bb/bb/bb/bb/bb",
			wantPass: true,
		},
		{
			ref:      "aa/aa/bb/bb/bb",
			wantPass: true,
		},
		{
			ref:      "a/a/a/a",
			wantPass: true,
		},
		{
			ref:      "a/a/a/a/",
			wantPass: false,
		},
		{
			ref:      "a//a/a",
			wantPass: false,
		},
		{
			ref:      "a",
			wantPass: true,
		},
		{
			ref:      "a/aa",
			wantPass: true,
		},
		{
			ref:      "a/aa/a",
			wantPass: true,
		},
		{
			ref:      "foo.com",
			wantPass: true,
		},
		{
			ref:      "foo.com/",
			wantPass: false,
		},
		{
			ref:      "foo.com:8080/bar",
			wantPass: true,
		},
		{
			ref:      "foo.com:http/bar",
			wantPass: false,
		},
		{
			ref:      "foo.com/bar",
			wantPass: true,
		},
		{
			ref:      "foo.com/bar/baz",
			wantPass: true,
		},
		{
			ref:      "localhost:8080/bar",
			wantPass: true,
		},
		{
			ref:      "sub-dom1.foo.com/bar/baz/quux",
			wantPass: true,
		},
		{
			ref:      "blog.foo.com/bar/baz",
			wantPass: true,
		},
		{
			ref:      "a^a",
			wantPass: false,
		},
		{
			ref:      "aa/asdf$$^/aa",
			wantPass: false,
		},
		{
			ref:      "asdf$$^/aa",
			wantPass: false,
		},
		{
			ref:      "aa-a/a",
			wantPass: true,
		},
		{
			ref:      strings.Repeat("a/", 128) + "a",
			wantPass: true,
		},
		{
			ref:      "a-/a/a/a",
			wantPass: false,
		},
		{
			ref:      "foo.com/a-/a/a",
			wantPass: false,
		},
		{
			ref:      "-foo/bar",
			wantPass: false,
		},
		{
			ref:      "foo/bar-",
			wantPass: false,
		},
		{
			ref:      "foo-/bar",
			wantPass: false,
		},
		{
			ref:      "foo/-bar",
			wantPass: false,
		},
		{
			ref:      "_foo/bar",
			wantPass: false,
		},
		{
			ref:      "foo_bar",
			wantPass: true,
		},
		{
			ref:      "foo_bar.com",
			wantPass: true,
		},
		{
			ref:      "foo_bar.com:8080",
			wantPass: true,
		},
		{
			ref:      "foo_bar.com:8080/app",
			wantPass: false,
		},
		{
			ref:      "foo.com/foo_bar",
			wantPass: true,
		},
		{
			ref:      "____/____",
			wantPass: false,
		},
		{
			ref:      "_docker/_docker",
			wantPass: false,
		},
		{
			ref:      "docker_/docker_",
			wantPass: false,
		},
		{
			ref:      "b.gcr.io/test.example.com/my-app",
			wantPass: true,
		},
		{
			ref:      "xn--n3h.com/myimage", // ☃.com in punycode
			wantPass: true,
		},
		{
			ref:      "xn--7o8h.com/myimage", // 🐳.com in punycode
			wantPass: true,
		},
		{
			ref:      "example.com/xn--7o8h.com/myimage", // 🐳.com in punycode
			wantPass: true,
		},
		{
			ref:      "example.com/some_separator__underscore/myimage",
			wantPass: true,
		},
		{
			ref:      "example.com/__underscore/myimage",
			wantPass: false,
		},
		{
			ref:      "example.com/..dots/myimage",
			wantPass: false,
		},
		{
			ref:      "example.com/.dots/myimage",
			wantPass: false,
		},
		{
			ref:      "example.com/nodouble..dots/myimage",
			wantPass: false,
		},
		{
			ref:      "example.com/nodouble..dots/myimage",
			wantPass: false,
		},
		{
			ref:      "docker./docker",
			wantPass: false,
		},
		{
			ref:      ".docker/docker",
			wantPass: false,
		},
		{
			ref:      "docker-/docker",
			wantPass: false,
		},
		{
			ref:      "-docker/docker",
			wantPass: false,
		},
		{
			ref:      "do..cker/docker",
			wantPass: false,
		},
		{
			ref:      "do__cker:8080/docker",
			wantPass: false,
		},
		{
			ref:      "do__cker/docker",
			wantPass: true,
		},
		{
			ref:      "b.gcr.io/test.example.com/my-app",
			wantPass: true,
		},
		{
			ref:      "registry.io/foo/project--id.module--name.ver---sion--name",
			wantPass: true,
		},
		{
			ref:      "Asdf.com/foo/bar", // uppercase character in hostname
			wantPass: true,
		},
		{
			ref:      "Foo/FarB", // uppercase characters in remote name
			wantPass: false,
		},
	}

	for i, tc := range testcases {
		r := NewDeploymentImageRefValid()
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got, _ := r.Exec(nilCtx, &api.Deployment{
				Spec: deployments.DeviceDeploymentSpec{
					Template: deployments.PodTemplate{
						Spec: v20190817Pod.PodConfig{
							Containers: []v20190817Pod.Container{
								{
									Image: tc.ref,
								},
							},
						},
					},
				},
			})
			got.SourceID = "test"
			got.Rule = r

			if tc.wantPass != got.Pass {
				t.Errorf("wanted pass=%v, got pass=%v", tc.wantPass, got.Pass)
			}
		})
	}
}
