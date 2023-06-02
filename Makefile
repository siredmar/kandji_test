DETECTED_OS := $(shell sh -c 'uname -s | tr "[:upper:]" "[:lower:]" 2>/dev/null || echo unknown')
DETECTED_ARCH := $(shell sh -c "uname -p | sed 's/x86_/amd/' 2>/dev/null || echo unknown")
DETECTED_PLATFORM := ${DETECTED_OS}/${DETECTED_ARCH}

# M1 platform override
ifeq ("$(DETECTED_PLATFORM)","darwin/arm")
define dockerBuild
	docker buildx build --platform=linux/arm64/v8 --build-arg SRC_BIN=./bin/gxctl-$1-$2 -t gxctl-$1-$2 .
endef
else
define dockerBuild
	docker buildx build --platform=$3 --build-arg SRC_BIN=./bin/gxctl-$1-$2 -t gxctl-$1-$2 .
endef
endif

HOST_OS ?= $(shell sh -c "go env GOOS 2>/dev/null || echo ${DETECTED_OS}")
HOST_ARCH ?= $(shell sh -c "go env GOARCH 2>/dev/null || echo ${DETECTED_ARCH}")
HOST_PLATFORM ?= ${HOST_OS}/${HOST_ARCH}

TARGET_OS ?= ${HOST_OS}
TARGET_ARCH ?= ${HOST_ARCH}
SRC_BIN ?= ./bin/gxctl-${TARGET_OS}-${TARGET_ARCH}

GIT_COMMIT := $(shell git rev-list -1 HEAD)
BUILDTIME := $(shell date)
VERSION ?= $(shell bin/version.sh)

GO_TOOLS := public.ecr.aws/gridx/base-images:golang-docker-dev-1.17.latest
GO_PROJECT := github.com/grid-x/gxctl

BRANCH := $(shell echo ${BUILDKITE_BRANCH} | sed 's/\//_/g')
IMAGE_TAG := ${BRANCH}.${BUILDKITE_BUILD_NUMBER}-${BUILDKITE_COMMIT}
GXCTL_BASE_URL := 108014196837.dkr.ecr.eu-central-1.amazonaws.com/gridx/gxctl
GXCTL_IMAGE_URL ?= ${GXCTL_BASE_URL}:${IMAGE_TAG}

DOCKER_RUN := docker run -e HOST_OS=${HOST_OS} -e HOST_ARCH=${HOST_ARCH} -e TARGET_OS=${TARGET_OS} -e TARGET_ARCH=${TARGET_ARCH} -e GXCTL_IMAGE_URL=${GXCTL_IMAGE_URL} --init -it --rm -v $$PWD:/go/src/${GO_PROJECT}:z -v /var/run/docker.sock:/var/run/docker.sock -w /go/src/${GO_PROJECT} ${GO_TOOLS} bash -c

define goBuild
	GOOS=$1 GOARCH=$2 CGO_ENABLED=0 go build -o ./bin/gxctl-$1-$2 -ldflags="-w -s -X 'github.com/grid-x/gxctl/internal/version.GitCommit=$(GIT_COMMIT)' -X 'github.com/grid-x/gxctl/internal/version.BuildTime=$(BUILDTIME)' -X 'github.com/grid-x/gxctl/internal/version.Version=$(VERSION)'" ./cmd/gxctl
endef

.PHONY: build bin/gxctl-* lint test ssh_config release docker_* ci_

build:
	$(call goBuild,${HOST_OS},${HOST_ARCH})

bin/gxctl-linux-amd64:
	$(call goBuild,linux,amd64)

bin/gxctl-linux-arm64:
	$(call goBuild,linux,arm64)

bin/gxctl-darwin-amd64:
	$(call goBuild,darwin,amd64)

bin/gxctl-darwin-arm64:
	$(call goBuild,darwin,arm64)

all: bin/gxctl-linux-amd64 bin/gxctl-linux-arm64 bin/gxctl-darwin-amd64 bin/gxctl-darwin-arm64

lint:
	golint -set_exit_status $(shell go list ./...)

test:
	go test -v $(shell go list ./...)

ssh_config:
	@jq -r 'to_entries[] | "Host \(.key)\n\((.value | to_entries | map("  \(.key) \(.value)")) | join("\n"))\n"' \
	pkg/action/ssh_config.json > \
	bin/ssh_config

docker: docker_${HOST_OS}_${HOST_ARCH}

docker_linux_amd64: ssh_config
	$(call dockerBuild,linux,amd64,linux/amd64)

docker_linux_arm64: ssh_config
	$(call dockerBuild,linux,arm64,linux/arm64/v8)

docker_all: docker_linux_amd64 docker_linux_arm64

docker_push:
	docker tag gxctl-linux-amd64 ${GXCTL_IMAGE_URL}-linux-amd64
	docker tag gxctl-linux-arm64 ${GXCTL_IMAGE_URL}-linux-arm64
	docker push ${GXCTL_IMAGE_URL}-linux-amd64
	docker push ${GXCTL_IMAGE_URL}-linux-arm64
	docker manifest create --amend ${GXCTL_IMAGE_URL} \
	  ${GXCTL_IMAGE_URL}-linux-amd64 \
	  ${GXCTL_IMAGE_URL}-linux-arm64
	docker manifest annotate ${GXCTL_IMAGE_URL} ${GXCTL_IMAGE_URL}-linux-amd64 --arch amd64 --os linux
	docker manifest annotate ${GXCTL_IMAGE_URL} ${GXCTL_IMAGE_URL}-linux-arm64 --arch arm64 --os linux --variant v8
	docker manifest push ${GXCTL_IMAGE_URL}

release:
	# TODO actually integrate publishing
	goreleaser release --skip-publish

ci_build:
	${DOCKER_RUN} "make bin/gxctl-linux-amd64"
	${DOCKER_RUN} "make bin/gxctl-linux-arm64"

ci_lint:
	${DOCKER_RUN} "make lint"

ci_test:
	${DOCKER_RUN} "make test"

ci_release:
	${DOCKER_RUN} "goreleaser release --skip-publish --skip-validate --rm-dist"
