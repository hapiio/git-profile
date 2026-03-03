.PHONY: build test lint clean install snapshot cover vet

BINARY   := git-profile
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE     := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS  := -ldflags "-s -w \
  -X main.version=$(VERSION) \
  -X main.commit=$(COMMIT) \
  -X main.date=$(DATE)"

## build: Compile the binary to ./git-profile
build:
	go build $(LDFLAGS) -o $(BINARY) .

## install: Install to $GOPATH/bin (or $GOBIN)
install:
	go install $(LDFLAGS) .

## test: Run all unit tests with race detection
test:
	go test -v -race -coverprofile=coverage.out ./...

## cover: Open coverage report in browser
cover: test
	go tool cover -html=coverage.out

## vet: Run go vet
vet:
	go vet ./...

## lint: Run golangci-lint (requires golangci-lint to be installed)
lint:
	golangci-lint run ./...

## snapshot: Build a GoReleaser snapshot (no publish)
snapshot:
	goreleaser release --snapshot --clean

## clean: Remove build artifacts
clean:
	rm -f $(BINARY) coverage.out

## help: Show this help
help:
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
