GO_FILES := $(shell find . -type f -name "*.go")
GO_BUILD := CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/manager -ldflags="-w -s" github.com/grid-x/ds-k8s/cmd/manager

GO_TOOLS := gridx/golang-tools:1.11
GO_PROJECT := github.com/grid-x/ds-k8s
DOCKER_RUN := docker run -it ${DOCKER_LINK} --rm -v $$PWD:/go/src/${GO_PROJECT} -w /go/src/${GO_PROJECT} -e BUILDKITE_BRANCH=${BUILDKITE_BRANCH} -e BUILDKITE_BUILD_NUMBER=${BUILDKITE_BUILD_NUMBER} -e BUILDKITE_COMMIT=${BUILDKITE_COMMIT}
GO_RUN := ${DOCKER_RUN} ${GO_TOOLS} bash -c

BASE_IMAGE_URL := 485611583707.dkr.ecr.eu-central-1.amazonaws.com/gridx/ds-k8s
BRANCH := $(shell echo ${BUILDKITE_BRANCH} | sed 's/\//_/g')
IMAGE_TAG := ${BRANCH}.${BUILDKITE_BUILD_NUMBER}-${BUILDKITE_COMMIT}
MANAGER_IMAGE_URL := ${BASE_IMAGE_URL}:${IMAGE_TAG}

# Image URL to use all building/pushing image targets
IMG ?= manager:latest

OPS_TOOLS := gridx/ops-tools:1.0.201901041341-linux-amd64
ifeq ($(CI),true)
	AWS_CREDS=-e AWS_ACCESS_KEY=$$TF_DS_PROD_AWS_ACCESS_KEY -e AWS_SECRET_KEY=$$TF_DS_PROD_AWS_SECRET_KEY
else
	AWS_CREDS=-v $$HOME/.aws:/root/.aws
endif

CLUSTER_NAME ?= ci-test-cluster-${BUILDKITE_BUILD_NUMBER}
EKSCTL_CREATE_COMMAND := eksctl create cluster --region=eu-central-1 --version=1.11 --name=${CLUSTER_NAME} --nodes=1 --node-type=t2.small --kubeconfig=./ci-test-cluster.yaml
EKSCTL_DELETE_COMMAND := eksctl delete cluster --region=eu-central-1 --name=${CLUSTER_NAME}
CREATE_EKS_CLUSTER := docker run -it --rm $(AWS_CREDS) -v $$PWD:/code -w /code ${OPS_TOOLS} bash -c "${EKSCTL_CREATE_COMMAND}"
DELETE_EKS_CLUSTER := docker run -it --rm $(AWS_CREDS) -v $$PWD:/code -w /code ${OPS_TOOLS} bash -c "${EKSCTL_DELETE_COMMAND}"
DEPLOY_CRD_EKS_CLUSTER := docker run -it --rm $(AWS_CREDS) -v $$PWD:/code -w /code ${OPS_TOOLS} bash -c "kubectl --kubeconfig=./ci-test-cluster.yaml apply -f config/crds"
DEPLOY_RBAC_EKS_CLUSTER := docker run -it --rm $(AWS_CREDS) -v $$PWD:/code -w /code ${OPS_TOOLS} bash -c "kubectl --kubeconfig=./ci-test-cluster.yaml apply -f config/rbac"
DOCKER_RUN_E2E := docker run -it $(AWS_CREDS) ${DOCKER_LINK} --rm -v $$PWD:/go/src/${GO_PROJECT} -w /go/src/${GO_PROJECT} -e BUILDKITE_BRANCH=${BUILDKITE_BRANCH} -e BUILDKITE_BUILD_NUMBER=${BUILDKITE_BUILD_NUMBER} -e BUILDKITE_COMMIT=${BUILDKITE_COMMIT}
GO_RUN_E2E := ${DOCKER_RUN_E2E} ${GO_TOOLS} bash -c

all: test manager

# Run tests
test: generate fmt vet manifests
	go test ./pkg/... ./cmd/... -coverprofile cover.out

test_e2e:
	# Setup IAM authenticator for working with EKS
	wget https://amazon-eks.s3-us-west-2.amazonaws.com/1.11.5/2018-12-06/bin/linux/amd64/aws-iam-authenticator
	mv aws-iam-authenticator /usr/local/bin/aws-iam-authenticator
	chmod +x /usr/local/bin/aws-iam-authenticator
	# Run e2e
	go test -timeout=2h -v ./test/e2e -kubeconfig $$PWD/ci-test-cluster.yaml -manager.image=${MANAGER_IMAGE_URL}

# Build manager binary
manager: generate fmt vet
	${GO_BUILD}

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

ci_docker:
	docker build -t ${IMG} -f Dockerfile .

ci_docker_push:
	docker tag ${IMG} ${MANAGER_IMAGE_URL}
	docker push ${MANAGER_IMAGE_URL}

ci_crd_deploy:
	${DEPLOY_CRD_EKS_CLUSTER}

ci_rbac_deploy:
	${DEPLOY_RBAC_EKS_CLUSTER}

ci_create_eks_cluster:
	${CREATE_EKS_CLUSTER}

ci_delete_eks_cluster:
	${DELETE_EKS_CLUSTER}

ci_docker_e2e: ci_docker_push ci_create_eks_cluster ci_crd_deploy ci_rbac_deploy
	${GO_RUN_E2E} "make test_e2e"
