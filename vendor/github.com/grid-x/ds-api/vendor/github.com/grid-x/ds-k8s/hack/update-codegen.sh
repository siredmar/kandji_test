#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ${GOPATH}/src/k8s.io/code-generator)}

VERSIONS="core:v1beta1 apps:v1beta1 batch:v1beta1 maintenance:v1beta1"

vendor/k8s.io/code-generator/generate-groups.sh all \
                                                github.com/grid-x/ds-k8s/pkg/client github.com/grid-x/ds-k8s/pkg/apis \
                                                "${VERSIONS}" \
                                                --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt
