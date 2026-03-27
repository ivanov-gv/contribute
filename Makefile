# Load .env if it exists (gitignored, copy from .env.example)
-include .env
export

BINARY    := gh-contribute
BUILD_DIR := bin

.PHONY: build test test-integration lint fmt tidy clean install

## build: compile the binary to bin/
build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/$(BINARY)

## install: install the binary to $GOPATH/bin
install:
	go install ./cmd/$(BINARY)

## test: run unit tests with race detector
test:
	go test -count=1 -race ./internal/...

## test-integration-local: run edge-case integration tests with mock server (no token needed)
test-integration-local:
	go test -count=1 -race ./test/integration/...

## test-integration: run integration tests against real GitHub API (requires GH_CONTRIBUTE_TOKEN)
test-integration:
	go test -tags integration -count=1 -race ./test/integration/...

## test-e2e: run E2E tests against real GitHub API (requires GH_CONTRIBUTE_TOKEN)
test-e2e:
	go test -tags integration -count=1 -race ./test/...

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## fmt: format all Go source files
fmt:
	gofmt -w ./...

## tidy: tidy and verify go modules
tidy:
	go mod tidy
	go mod verify

## clean: remove build artifacts
clean:
	rm -rf $(BUILD_DIR)


release-build:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o build/gh-contribute-windows-amd64.exe ./cmd/gh-contribute/
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/gh-contribute-linux-amd64 ./cmd/gh-contribute/
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o build/gh-contribute-darwin-amd64 ./cmd/gh-contribute/
    # git tag v0.0.0
    # git push origin v0.0.0
    # gh release create v0.0.0 ./build/*amd64*
