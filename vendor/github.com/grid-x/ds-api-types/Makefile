GO_TOOLS := public.ecr.aws/gridx/base-images:golang-dev-1.17.latest
GO_PROJECT := github.com/grid-x/ds-api-types
DOCKER_RUN := docker run --init -it --rm -v $$PWD:/go/src/${GO_PROJECT} -w /go/src/${GO_PROJECT}
GO_RUN := ${DOCKER_RUN} ${GO_TOOLS} bash -c

lint:
	golint -set_exit_status $(shell go list ./...)

test:
	go test -v $(shell go list ./...)

build:
	CGO_ENABLED=0 go build -ldflags="-w -s"

ci_lint:
	${GO_RUN} "make lint"

ci_test:
	${GO_RUN} "make test"

ci_build:
	${GO_RUN} "make build"
