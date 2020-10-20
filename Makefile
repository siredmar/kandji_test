GIT_COMMIT := $(shell git rev-list -1 HEAD)
BUILDTIME := $(shell date)
VERSION := $(shell cat VERSION)
GO_BUILD := CGO_ENABLED=0 go build -o bin/gxctl -ldflags=\"-w -s -X 'github.com/grid-x/gxctl/internal/version.GitCommit=$(GIT_COMMIT)' -X 'github.com/grid-x/gxctl/internal/version.BuildTime=$(BUILDTIME)' -X 'github.com/grid-x/gxctl/internal/version.Version=$(VERSION)'\" ./cmd/gxctl
GO_TOOLS := 108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/base-images:golang-dev-1.15.latest
GO_PROJECT := github.com/grid-x/gxctl
DOCKER_RUN := docker run -it --rm -v $$PWD:/go/src/${GO_PROJECT} -w /go/src/${GO_PROJECT}
GO_RUN := ${DOCKER_RUN} ${GO_TOOLS} bash -c

BRANCH := $(shell echo ${BUILDKITE_BRANCH} | sed 's/\//_/g')
IMAGE_TAG := ${BRANCH}.${BUILDKITE_BUILD_NUMBER}-${BUILDKITE_COMMIT}
GXCTL_BASE_URL := 108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/gxctl
GXCTL_IMAGE_URL := ${GXCTL_BASE_URL}:${IMAGE_TAG}
GXCTL_IMAGE_URL_RELEASE := ${GXCTL_BASE_URL}:${VERSION}

.PHONY: test

lint:
	golint -set_exit_status $(shell go list ./...)

test:
	go test -v $(shell go list ./...)

build: 
	bash -c "${GO_BUILD}"

ci_lint:
	${GO_RUN} "make lint"

ci_build:
	${DOCKER_RUN} ${GO_TOOLS} bash -c "${GO_BUILD}"

ci_test:
	${GO_RUN} "make test"

docker:
	docker build -t gxctl -f Dockerfile .

docker_push:
	docker tag gxctl ${GXCTL_IMAGE_URL}-linux-amd64
	docker push ${GXCTL_IMAGE_URL}-linux-amd64

docker_push_release:
	docker tag gxctl ${GXCTL_IMAGE_URL}-linux-amd64
	docker tag gxctl ${GXCTL_IMAGE_URL_RELEASE}-linux-amd64
	docker push ${GXCTL_IMAGE_URL}-linux-amd64
	docker push ${GXCTL_IMAGE_URL_RELEASE}-linux-amd64
