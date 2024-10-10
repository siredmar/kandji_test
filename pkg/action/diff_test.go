package action

import (
	_ "embed"
	"os"
	"os/exec"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/google/go-cmp/cmp"

	"github.com/grid-x/gxctl/pkg/api"
)

var testYaml = []byte(`
name: Go
version: 1.18
description: "Go is an open-source programming language."

dependencies:
  - name: yaml.v3
    version: 3.0.0
  - name: fmt
    version: 1.10.0

features:
  concurrency:
    description: "Built-in support for concurrent programming."
    examples:
      - "go run main.go"
      - "go test ./..."

config:
  timeout: 30
  options:
    logging: true
    debug: false

servers:
  - address: "127.0.0.1"
    port: 8080
  - address: "192.168.1.1"
    port: 8081

metadata:
  created: "2023-10-09"
  authors:
    - name: Alice
      email: alice@example.com
    - name: Bob
      email: bob@example.com

build:
  steps:
    - checkout:
        path: /path/to/repo
    - test:
        command: "go test ./..."
`)

func Test_flatten(t *testing.T) {
	deploymentYamlMap := make(map[string]any)
	if err := yaml.Unmarshal(testYaml, &deploymentYamlMap); err != nil {
		t.Errorf("failed to unmarshal yaml into a map: %v", err)
	}

	flatMap := make(map[string]any)
	flatten("", deploymentYamlMap, ".", flatMap)

	want := []yamlAsKV{
		{key: "build.steps[0].checkout.path", val: "/path/to/repo"},
		{key: "build.steps[1].test.command", val: "go test ./..."},
		{key: "config.options.debug", val: false},
		{key: "config.options.logging", val: true},
		{key: "config.timeout", val: float64(30)},
		{key: "dependencies[0].name", val: "yaml.v3"},
		{key: "dependencies[0].version", val: "3.0.0"},
		{key: "dependencies[1].name", val: "fmt"},
		{key: "dependencies[1].version", val: "1.10.0"},
		{key: "description", val: "Go is an open-source programming language."},
		{key: "features.concurrency.description", val: "Built-in support for concurrent programming."},
		{key: "features.concurrency.examples[0]", val: "go run main.go"},
		{key: "features.concurrency.examples[1]", val: "go test ./..."},
		{key: "metadata.authors[0].email", val: "alice@example.com"},
		{key: "metadata.authors[0].name", val: "Alice"},
		{key: "metadata.authors[1].email", val: "bob@example.com"},
		{key: "metadata.authors[1].name", val: "Bob"},
		{key: "metadata.created", val: "2023-10-09"},
		{key: "name", val: "Go"},
		{key: "servers[0].address", val: "127.0.0.1"},
		{key: "servers[0].port", val: float64(8080)},
		{key: "servers[1].address", val: "192.168.1.1"},
		{key: "servers[1].port", val: float64(8081)},
		{key: "version", val: float64(1.18)},
	}

	// this part is required to make the test results consistent (flatten iterates map in a random order)
	sortedList := yamlToSortedList(flatMap)

	if diff := cmp.Diff(want, sortedList, cmp.AllowUnexported(want[0])); diff != "" {
		t.Errorf("flatten mismatch (-want +got):\n%s", diff)
	}
}

//go:embed testdata/deployment1.yaml
var deployment1 []byte

//go:embed testdata/deployment2.yaml
var deployment2 []byte

func Test_diff(t *testing.T) {
	var d1, d2 api.Deployment
	if err := yaml.Unmarshal(deployment1, &d1); err != nil {
		t.Errorf("failed to unmarshal bytes into a deployment: %v", err)
	}
	if err := yaml.Unmarshal(deployment2, &d2); err != nil {
		t.Errorf("failed to unmarshal bytes into a deployment: %v", err)
	}

	f1, err := writeObjectToDiffFile(d1)
	defer os.Remove(f1)
	if err != nil {
		t.Errorf("failed to write an object to a diff file")
	}

	f2, err := writeObjectToDiffFile(d2)
	defer os.Remove(f2)
	if err != nil {
		t.Errorf("failed to write an object to a diff file")
	}

	output, err := exec.Command("diff", f1, f2).Output()
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			// this is just an exit code error, no worries
		default: // couldnt run diff
			t.Errorf("couldn't run diff: %v", err)
		}
	}

	want := `7c7
< spec.canary.template.spec.containers[0].image: 108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/gridbox-integrations:weekly_2024-09-24-2024_09_24-1063-917e644
---
> spec.canary.template.spec.containers[0].image: 108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/gridbox-integrations:weekly_2024-09-24-2024_09_24-1063-917e646
24c24
< spec.template.spec.containers[0].image: 108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/gridbox-integrations:weekly_2024-09-24-2024_09_24-1063-917e643
---
> spec.template.spec.containers[0].image: 108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/gridbox-integrations:weekly_2024-09-24-2024_09_24-1063-917e645
`

	if diff := cmp.Diff(want, string(output)); diff != "" {
		t.Errorf("diff mistmatch: (-want +got):\n%s", diff)
	}
}
