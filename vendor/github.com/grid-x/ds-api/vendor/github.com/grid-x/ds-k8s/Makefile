GO_FILES := $(shell find . -type f -name "*.go")
GO_BUILD := CGO_ENABLED=0 go build -ldflags="-w -s"
GO_TOOLS := gridx/golang-tools:1.11
GO_PROJECT := github.com/grid-x/ds-k8s
DOCKER_RUN := docker run -it ${DOCKER_LINK} --rm -v $$PWD:/go/src/${GO_PROJECT} -w /go/src/${GO_PROJECT}
GO_RUN := ${DOCKER_RUN} ${GO_TOOLS} bash -c

BASE_IMAGE_URL :=${AWS_ACCOUNT_ID}.dkr.ecr.eu-central-1.amazonaws.com/gridx
BRANCH := $(shell echo ${BUILDKITE_BRANCH} | sed 's/\//_/g')
IMAGE_TAG := ${BRANCH}.${BUILDKITE_BUILD_NUMBER}-${BUILDKITE_COMMIT}

# Image URL to use all building/pushing image targets
IMG ?= controller:latest

all: test manager

# Run tests
test: generate fmt vet manifests
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager github.com/grid-x/ds-k8s/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crds
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

lint:
	golint -set_exit_status ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...

# Build the docker image
docker-build: test
	docker build . -t ${IMG}
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml

# Push the docker image
docker-push:
	docker push ${IMG}

update_codgen:
	hack/update-codegen.sh

verify_codegen:
	hack/verify-codegen.sh

kubebuilder:
	./scripts/install_kubebuilder.sh

deps: kubebuilder

ci_test:
	${GO_RUN} "make deps && make test"

ci_lint:
	${GO_RUN} "make lint"

ci_manager:
	${GO_RUN} "make manager"

ci_verify_codegen:
	${GO_RUN} "make generate && make verify_codegen"
