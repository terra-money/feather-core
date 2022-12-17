#! /usr/bin/make -f

# Project variables.
PROJECT_NAME = feather-based
DATE := $(shell date '+%Y-%m-%dT%H:%M:%S')
HEAD = $(shell git rev-parse HEAD)
LD_FLAGS = 
BUILD_FLAGS = -mod=readonly -ldflags='$(LD_FLAGS)'
BUILD_FOLDER = ./dist

## install: Install the binary.
install:
	@echo Installing $(PROJECT_NAME)...
	@go install $(BUILD_FLAGS) ./...
	@feather-based --version

## build: Build the binary.
build:
	@echo Building $(PROJECT_NAME)...
	@-mkdir -p $(BUILD_FOLDER) 2> /dev/null
	@go build $(BUILD_FLAGS) -o $(BUILD_FOLDER) ./...

## clean: Clean build files. Also runs `go clean` internally.
clean:
	@echo Cleaning build cache...
	@-rm -rf $(BUILD_FOLDER) 2> /dev/null
	@go clean ./...

.PHONY: install build mocks clean

## govet: Run go vet.
govet:
	@echo Running go vet...
	@go vet ./...

## govulncheck: Run govulncheck
govulncheck:
	@echo Running govulncheck...
	@go run golang.org/x/vuln/cmd/govulncheck ./...

## format: Install and run goimports and gofumpt
format:
	@echo Formatting...
	@go run mvdan.cc/gofumpt -w .
	@go run golang.org/x/tools/cmd/goimports -w -local github.com/terra-money/feather-toolkit .

## lint: Run Golang CI Lint.
lint:
	@echo Running gocilint...
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --out-format=tab --issues-exit-code=0

.PHONY: govet format lint

## test-unit: Run the unit tests.
test-unit:
	@echo Running unit tests...
	@go test -race -failfast -v ./...

.PHONY: test-unit test

help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECT_NAME)", or just run 'make' for install"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

.PHONY: help

.DEFAULT_GOAL := install