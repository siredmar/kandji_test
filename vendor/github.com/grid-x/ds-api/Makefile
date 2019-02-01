GO_FILES := $(shell find . -type f -name "*.go")
GO_BUILD := CGO_ENABLED=0 go build -ldflags="-w -s"
GO_TOOLS := gridx/golang-tools:1.11
GO_PROJECT := github.com/grid-x/ds-api
DOCKER_RUN := docker run -it ${DOCKER_LINK} --rm -v $$PWD:/go/src/${GO_PROJECT} -w /go/src/${GO_PROJECT}
GO_RUN := ${DOCKER_RUN} ${GO_TOOLS} bash -c
DOCKER_TEST_NET := test_${BUILDKITE_BUILD_NUMBER}_${BUILDKITE_COMMIT}
POSTGRES_TEST_NAME := postgres_test_${BUILDKITE_BUILD_NUMBER}_${BUILDKITE_COMMIT}

BRANCH := $(shell echo ${BUILDKITE_BRANCH} | sed 's/\//_/g')
IMAGE_TAG := ${BRANCH}.${BUILDKITE_BUILD_NUMBER}-${BUILDKITE_COMMIT}
DS_API_BASE_URL := 485611583707.dkr.ecr.eu-central-1.amazonaws.com/gridx/ds-api
DS_API_IMAGE_URL := ${DS_API_BASE_URL}:${IMAGE_TAG}

all: bin/ds-dev

cmd/ds-dev/templates_gen.go: cmd/ds-dev/templates.go cmd/ds-dev/templates/*.tmpl
	go generate cmd/ds-dev/templates.go


.PHONY: pkg/postgres/migrations/sql/*.sql
pkg/postgres/migrations/sql/*.sql:

pkg/postgres/migrations/migrations_gen.go: pkg/postgres/migrations/migrations.go pkg/postgres/migrations/sql/*.sql
	go generate pkg/postgres/migrations/migrations.go

bin/ds-dev: cmd/ds-dev/*.go cmd/ds-dev/templates_gen.go
	go build -o bin/ds-dev ./cmd/ds-dev

bin/server: cmd/server/*.go pkg/postgres/migrations/migrations_gen.go cmd/server/management_routes_gen.go cmd/server/device_routes_gen.go
	go build -o bin/server ./cmd/server

bin/server.linux: cmd/server/*.go pkg/postgres/migrations/migrations_gen.go cmd/server/management_routes_gen.go cmd/server/device_routes_gen.go
	GOOS=linux go build -o bin/server ./cmd/server

.PHONY: test_postgres
test_postgres: pkg/postgres/*.go pkg/postgres/migrations/migrations_gen.go
	go test -v ./pkg/postgres -postgres.url $$POSTGRES_URL

.PHONY: test
test: cmd/ds-dev/templates_gen.go  test_postgres cmd/server/device_routes_gen.go cmd/server/management_routes_gen.go
	go test -v $(shell go list ./... | grep -v "postgres" | grep -v "e2e")

.PHONY: lint
lint:
# for now we ignore ds-dev tool
	golint -set_exit_status $(shell go list ./... | grep -v "ds-dev\|auth\/device")

cmd/server/management_routes_gen.go: bin/ds-dev api/management/**/**/*.go
	bin/ds-dev routes -t management -o ./cmd/server/management_routes_gen.go

cmd/server/device_routes_gen.go: bin/ds-dev api/device/**/**/*.go
	bin/ds-dev routes -t device -o ./cmd/server/device_routes_gen.go

ci_deps:
	go get -v -u github.com/jteeuwen/go-bindata/...

ci_test:
	docker network create ${DOCKER_TEST_NET}
	docker run -d --network ${DOCKER_TEST_NET} -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=pw --name ${POSTGRES_TEST_NAME} postgres:10
	${DOCKER_RUN} --network ${DOCKER_TEST_NET} ${GO_TOOLS} bash -c "export GOPATH=/go && make ci_deps && POSTGRES_URL=postgres://postgres:pw@${POSTGRES_TEST_NAME}:5432/postgres?sslmode=disable make test"; \
  status=$$?; \
	docker stop ${POSTGRES_TEST_NAME} && docker rm ${POSTGRES_TEST_NAME}; \
	docker network rm ${DOCKER_TEST_NET}; \
  exit $$status

ci_lint:
	${GO_RUN} "make lint"

ci_ds_dev:
	${GO_RUN} "make ci_deps && make bin/ds-dev"

ci_server:
	${GO_RUN} "make ci_deps && make bin/server.linux"

.PHONY: test_routes
test_routes: cmd/server/management_routes_gen.go cmd/server/device_routes_gen.go
	git diff --quiet

.PHONY: ci_test_routes
ci_test_routes:
	${GO_RUN} "make ci_deps && make test_routes"

ci_docker:
	docker build -t ds-api -f Dockerfile .

ci_docker_push:
	docker tag ds-api ${DS_API_IMAGE_URL}
	docker push ${DS_API_IMAGE_URL}

test_e2e:
	minikube start
	docker run -d -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=pw -e POSTGRES_DB=testdb --name testdb postgres:10
	eval $(minikube docker-env)
	export AWS_PROFILE=ds-prod
	eval $(aws ecr get-login --no-include-email)
	docker pull 485611583707.dkr.ecr.eu-central-1.amazonaws.com/gridx/ds-k8s:master.96-55f04b2eaa1f1b7397c55863402cedbce76eb5e3
	make bin/server
	kubectl apply -f test/e2e/k8s-manager
	bin/server --k8s.config ~/.kube/config --auth.rsa-key=./test-certs/jwt.pem --auth.dev-rsa-key=./test-certs/jwt-key.pem --postgres.url="postgres://postgres:pw@192.168.168.55:5432/testdb?sslmode=disable" --postgres.migrate
