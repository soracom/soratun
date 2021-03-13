GO=go
BIN=soratun
SRC=$(shell find . -type f -name '*.go')
TEST_CONTAINER_NAME=soratun-test
TEST_CONTAINER_RESOURCE=$(TEST_CONTAINER_NAME):latest
GOLANGCI_VERSION=1.40.1

$(BIN): $(SRC) go.mod
	CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -trimpath ./cmd/soratun

snapshot: json-schema-docs
	which goreleaser && goreleaser --snapshot --skip-publish --rm-dist

clean:
	rm -fr $(BIN) dist

gen:
	go generate ./...

integration-test-container:
	docker build . -f ./devtools/wg_integ_test/Dockerfile -t $(TEST_CONTAINER_RESOURCE)

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
	docker exec -it soratun-test bash -c 'mkdir -p /dev/net ; mknod /dev/net/tun c 10 200 ; chmod 600 /dev/net/tun; WG_INTEG_TEST=enabled go test -v ./...'

clean-integration-test-container:
	docker container stop $(TEST_CONTAINER_NAME)
	docker image rm $(TEST_CONTAINER_RESOURCE)

json-schema-docs:
	which json-schema-docs && json-schema-docs -schema docs/schema/soratun-config.en.schema.json > docs/config.en.md
	which json-schema-docs && json-schema-docs -schema docs/schema/soratun-config.ja.schema.json > docs/config.ja.md

install-dev-deps:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
		| sh -s -- -b $(shell go env GOPATH)/bin v$(GOLANGCI_VERSION)

lint:
	@golangci-lint run ./...

test:
	@go test .

# FIXME: run ci tests on GH Actions after publishing. (ref. PR#32)
ci: install-dev-deps lint test

.PHONY: clean clean-integration-test-container json-schema-docs install-dev-deps lint test ci
