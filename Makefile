GIT_COMMIT := $(shell git rev-list -1 HEAD)
BUILDTIME := $(shell date)
VERSION ?= $(shell bin/version.sh)
# ignore Darwin host - docker needs linux binary
HOST_GOOS ?= $(shell go env GOOS)
HOST_GOARCH ?= $(shell go env GOARCH)
TARGET_GOOS ?= ${HOST_GOOS}
TARGET_GOARCH ?= ${HOST_GOARCH}
GO_BUILD := GOOS=${TARGET_GOOS} GOARCH=${TARGET_GOARCH} CGO_ENABLED=0 go build -o bin/gxctl-${TARGET_GOOS}-${TARGET_GOARCH} -ldflags=\"-w -s -X 'github.com/grid-x/gxctl/internal/version.GitCommit=$(GIT_COMMIT)' -X 'github.com/grid-x/gxctl/internal/version.BuildTime=$(BUILDTIME)' -X 'github.com/grid-x/gxctl/internal/version.Version=$(VERSION)'\" ./cmd/gxctl
# specific revision for goreleaser
GO_TOOLS := public.ecr.aws/gridx/base-images:golang-docker-dev-1.17.latest
GO_PROJECT := github.com/grid-x/gxctl

BRANCH := $(shell echo ${BUILDKITE_BRANCH} | sed 's/\//_/g')
IMAGE_TAG := ${BRANCH}.${BUILDKITE_BUILD_NUMBER}-${BUILDKITE_COMMIT}
GXCTL_BASE_URL := 108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/gxctl
GXCTL_IMAGE_URL ?= ${GXCTL_BASE_URL}:${IMAGE_TAG}

DOCKER_RUN := docker run -e HOST_GOOS=${HOST_GOOS} -e HOST_GOARCH=${HOST_GOARCH} -e TARGET_GOOS=${TARGET_GOOS} -e TARGET_GOARCH=${TARGET_GOARCH} -e GXCTL_IMAGE_URL=${GXCTL_IMAGE_URL} --init -it --rm -v $$PWD:/go/src/${GO_PROJECT}:z -v /var/run/docker.sock:/var/run/docker.sock -w /go/src/${GO_PROJECT}
GO_RUN := ${DOCKER_RUN} ${GO_TOOLS} bash -c

.PHONY: test

lint:
	golint -set_exit_status $(shell go list ./...)

test:
	go test -v $(shell go list ./...)

build: 
	bash -c "${GO_BUILD}"

release:
	# TODO actually integrate publishing
	goreleaser release --skip-publish

ci_lint:
	${GO_RUN} "make lint"

ci_build:
	${GO_RUN} "make build"

ci_test:
	${GO_RUN} "make test"

ci_release:
	${GO_RUN} "goreleaser release --skip-publish --skip-validate --rm-dist"

docker:
	docker build -t gxctl-${TARGET_GOOS}-${TARGET_GOARCH} -f Dockerfile \
		--build-arg HOST_GOOS=${HOST_GOOS} \
		--build-arg HOST_GOARCH=${HOST_GOARCH} \
		--build-arg TARGET_GOOS=${TARGET_GOOS} \
		--build-arg TARGET_GOARCH=${TARGET_GOARCH} \
		.

docker_push:
	docker tag gxctl-linux-amd64 ${GXCTL_IMAGE_URL}-linux-amd64
	docker tag gxctl-linux-arm64 ${GXCTL_IMAGE_URL}-linux-arm64
	docker push ${GXCTL_IMAGE_URL}-linux-amd64
	docker push ${GXCTL_IMAGE_URL}-linux-arm64
	docker manifest create ${GXCTL_IMAGE_URL} \
	  ${GXCTL_IMAGE_URL}-linux-amd64 \
	  ${GXCTL_IMAGE_URL}-linux-arm64
	docker manifest push ${GXCTL_IMAGE_URL}
