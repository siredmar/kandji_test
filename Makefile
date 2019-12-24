GIT_COMMIT := $(shell git rev-list -1 HEAD)
BUILDTIME := $(shell date)
GO_BUILD := CGO_ENABLED=0 go build -ldflags=\"-w -s -X github.com/grid-x/gxctl/cmd.gitCommit=$(GIT_COMMIT) -X 'github.com/grid-x/gxctl/cmd.buildTime=$(BUILDTIME)'\"
GO_TOOLS := gridx/golang-dev:1.13.latest-linux-amd64
GO_PROJECT := github.com/grid-x/gxctl
DOCKER_RUN := docker run -it --rm -v $$PWD:/go/src/${GO_PROJECT} -w /go/src/${GO_PROJECT}
GO_RUN := ${DOCKER_RUN} ${GO_TOOLS} bash -c

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
