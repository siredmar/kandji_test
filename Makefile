GO_BUILD := CGO_ENABLED=0 go build -ldflags="-w -s"
GO_TOOLS := gridx/golang-tools:1.11
GO_PROJECT := github.com/grid-x/gxctl
DOCKER_RUN := docker run -it --rm -v $$PWD:/go/src/${GO_PROJECT} -w /go/src/${GO_PROJECT}
GO_RUN := ${DOCKER_RUN} ${GO_TOOLS} bash -c

lint:
	golint -set_exit_status $(shell go list ./...)

build: 
	${GO_BUILD}

ci_lint:
	${GO_RUN} "make lint"

ci_build:
	${DOCKER_RUN} ${GO_TOOLS} bash -c "${GO_BUILD}"