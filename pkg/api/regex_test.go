package api

import (
	"strings"
	"testing"
)

type regexpValid struct {
	input string
	match bool
}

type regexpImageName struct {
	input   string
	result  string
	wantErr bool
}

func TestIsDockerImageValid(t *testing.T) {
	testcases := []regexpValid{
		{
			input: "",
			match: false,
		},
		{
			input: "short",
			match: true,
		},
		{
			input: "simple/name",
			match: true,
		},
		{
			input: "library/ubuntu",
			match: true,
		},
		{
			input: "docker/stevvooe/app",
			match: true,
		},
		{
			input: "aa/aa/aa/aa/aa/aa/aa/aa/aa/bb/bb/bb/bb/bb/bb",
			match: true,
		},
		{
			input: "aa/aa/bb/bb/bb",
			match: true,
		},
		{
			input: "a/a/a/a",
			match: true,
		},
		{
			input: "a/a/a/a/",
			match: false,
		},
		{
			input: "a//a/a",
			match: false,
		},
		{
			input: "a",
			match: true,
		},
		{
			input: "a/aa",
			match: true,
		},
		{
			input: "a/aa/a",
			match: true,
		},
		{
			input: "foo.com",
			match: true,
		},
		{
			input: "foo.com/",
			match: false,
		},
		{
			input: "foo.com:8080/bar",
			match: true,
		},
		{
			input: "foo.com:http/bar",
			match: false,
		},
		{
			input: "foo.com/bar",
			match: true,
		},
		{
			input: "foo.com/bar/baz",
			match: true,
		},
		{
			input: "localhost:8080/bar",
			match: true,
		},
		{
			input: "sub-dom1.foo.com/bar/baz/quux",
			match: true,
		},
		{
			input: "blog.foo.com/bar/baz",
			match: true,
		},
		{
			input: "a^a",
			match: false,
		},
		{
			input: "aa/asdf$$^/aa",
			match: false,
		},
		{
			input: "asdf$$^/aa",
			match: false,
		},
		{
			input: "aa-a/a",
			match: true,
		},
		{
			input: strings.Repeat("a/", 128) + "a",
			match: true,
		},
		{
			input: "a-/a/a/a",
			match: false,
		},
		{
			input: "foo.com/a-/a/a",
			match: false,
		},
		{
			input: "-foo/bar",
			match: false,
		},
		{
			input: "foo/bar-",
			match: false,
		},
		{
			input: "foo-/bar",
			match: false,
		},
		{
			input: "foo/-bar",
			match: false,
		},
		{
			input: "_foo/bar",
			match: false,
		},
		{
			input: "foo_bar",
			match: true,
		},
		{
			input: "foo_bar.com",
			match: true,
		},
		{
			input: "foo_bar.com:8080",
			match: true,
		},
		{
			input: "foo_bar.com:8080/app",
			match: false,
		},
		{
			input: "foo.com/foo_bar",
			match: true,
		},
		{
			input: "____/____",
			match: false,
		},
		{
			input: "_docker/_docker",
			match: false,
		},
		{
			input: "docker_/docker_",
			match: false,
		},
		{
			input: "b.gcr.io/test.example.com/my-app",
			match: true,
		},
		{
			input: "xn--n3h.com/myimage", // ☃.com in punycode
			match: true,
		},
		{
			input: "xn--7o8h.com/myimage", // 🐳.com in punycode
			match: true,
		},
		{
			input: "example.com/xn--7o8h.com/myimage", // 🐳.com in punycode
			match: true,
		},
		{
			input: "example.com/some_separator__underscore/myimage",
			match: true,
		},
		{
			input: "example.com/__underscore/myimage",
			match: false,
		},
		{
			input: "example.com/..dots/myimage",
			match: false,
		},
		{
			input: "example.com/.dots/myimage",
			match: false,
		},
		{
			input: "example.com/nodouble..dots/myimage",
			match: false,
		},
		{
			input: "example.com/nodouble..dots/myimage",
			match: false,
		},
		{
			input: "docker./docker",
			match: false,
		},
		{
			input: ".docker/docker",
			match: false,
		},
		{
			input: "docker-/docker",
			match: false,
		},
		{
			input: "-docker/docker",
			match: false,
		},
		{
			input: "do..cker/docker",
			match: false,
		},
		{
			input: "do__cker:8080/docker",
			match: false,
		},
		{
			input: "do__cker/docker",
			match: true,
		},
		{
			input: "b.gcr.io/test.example.com/my-app",
			match: true,
		},
		{
			input: "registry.io/foo/project--id.module--name.ver---sion--name",
			match: true,
		},
		{
			input: "Asdf.com/foo/bar", // uppercase character in hostname
			match: true,
		},
		{
			input: "Foo/FarB", // uppercase characters in remote name
			match: false,
		},
	}
	for i := range testcases {
		res := IsDockerImageValid(testcases[i].input)
		if res != testcases[i].match {
			t.Errorf("Result does not match. Expected: %t, Got: %t for string: %s", testcases[i].match, res, testcases[i].input)
		}
	}
}

func TestGetDockerImageName(t *testing.T) {
	testcases := []regexpImageName{
		{
			input:   "gridx/test",
			result:  "gridx-test-latest",
			wantErr: false,
		},
		{
			input:   "gridx/test:123456",
			result:  "gridx-test-123456",
			wantErr: false,
		},
		{
			input:   "gridx/test/test2:123456",
			result:  "gridx-test-test2-123456",
			wantErr: false,
		},
	}

	for i := range testcases {
		res, err := GetDockerImageName(testcases[i].input)
		if err != nil && !testcases[i].wantErr {
			t.Error("Did not expect an error")
		}
		if err == nil && testcases[i].wantErr {
			t.Error("Did expect an error")
		}

		if res != testcases[i].result {
			t.Errorf("Result does not match. Expected: %s, Got: %s", testcases[i].result, res)
		}
	}
}
