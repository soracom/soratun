GO=go
BIN=soratun
SRC=$(shell find . -type f -name '*.go')
TEST_CONTAINER_NAME=soratun-test
TEST_CONTAINER_RESOURCE=$(TEST_CONTAINER_NAME):latest

GOLANGCI_VERSION=1.61.0
GOIMPORTS_VERSION=0.12.0
MOCKGEN_VERSION=0.2.0
JSON_SCHEMA_DOCS_VERSION=0.2.1
GORELEASER_VERSION=1.20.0

check: fmt-check test lint vet
check-ci: fmt-check test vet

$(BIN): $(SRC) go.mod
	CGO_ENABLED=0 $(GO) build -ldflags='-s -w -X "github.com/soracom/soratun/internal.Revision=$(shell git rev-parse HEAD)" -X "github.com/soracom/soratun/internal.Version=$(shell git describe --tags $$(git rev-list --tags --max-count=1))"' -trimpath ./cmd/soratun

snapshot: json-schema-docs
	which goreleaser && goreleaser --snapshot --skip-publish --clean

clean:
	rm -fr $(BIN) dist

gen:
	go generate ./...

test:
	go test -v ./...

integration-test-container:
	docker build . -f ./devtools/wg_integ_test/Dockerfile -t $(TEST_CONTAINER_RESOURCE)
	docker tag $(TEST_CONTAINER_RESOURCE) ghcr.io/soracom/soratun/$(TEST_CONTAINER_RESOURCE)

run-integration-test-container: $(BIN) gen
	docker run -d \
		--name=$(TEST_CONTAINER_NAME) \
		--rm \
		--cap-add=NET_ADMIN \
		--cap-add=SYS_MODULE \
		-e PUID=1000 \
		-e PGID=1000 \
		-e TZ=UTC \
		-v $(PWD):/soratun \
		-w /soratun \
		--sysctl="net.ipv4.conf.all.src_valid_mark=1" \
		$(TEST_CONTAINER_RESOURCE)

integration-test:
	docker exec -it $(TEST_CONTAINER_NAME) bash -c 'mkdir -p /dev/net ; mknod /dev/net/tun c 10 200 ; chmod 600 /dev/net/tun; WG_INTEG_TEST=enabled go test -v ./...'

clean-integration-test-container:
	docker container stop $(TEST_CONTAINER_NAME)
	docker image rm $(TEST_CONTAINER_RESOURCE)

json-schema-docs:
	which json-schema-docs && json-schema-docs -schema docs/schema/soratun-config.en.schema.json > docs/config.en.md
	which json-schema-docs && json-schema-docs -schema docs/schema/soratun-config.ja.schema.json > docs/config.ja.md

install-dev-deps:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
		| sh -s -- -b $(shell go env GOPATH)/bin v$(GOLANGCI_VERSION) \
		&& go install golang.org/x/tools/cmd/goimports@v$(GOIMPORTS_VERSION) \
		&& go install go.uber.org/mock/mockgen@v$(MOCKGEN_VERSION) \
		&& go install github.com/marcusolsson/json-schema-docs@v$(JSON_SCHEMA_DOCS_VERSION) \
		&& go install github.com/goreleaser/goreleaser@v$(GORELEASER_VERSION)

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

vet:
	go vet ./...

fmt-check:
	goimports -l *.go **/*.go | grep [^*][.]go$$; \
	EXIT_CODE=$$?; \
	if [ $$EXIT_CODE -eq 0 ]; then exit 1; fi \

fmt:
	goimports -w *.go **/*.go

fix:
	$(MAKE) fmt
	$(MAKE) lint-fix

github-docker-login:
ifndef DOCKER_USER
	@echo "[error] \$$DOCKER_USER must be specified"
	@exit 1
endif
ifndef DOCKER_PSWD_FILE
	@echo "[error] \$$DOCKER_PSWD_FILE must be specified"
	@exit 1
endif
	cat $(DOCKER_PSWD_FILE) | docker login ghcr.io --username $(DOCKER_USER) --password-stdin

test-docker-container-push: github-docker-login
	docker push ghcr.io/soracom/soratun/$(TEST_CONTAINER_RESOURCE)

.PHONY: clean integration-test-container clean-integration-test-container run-integration-test-container json-schema-docs install-dev-deps lint test check check-ci vet fmt fmt-check github-docker-login test-docker-container test-docker-push

