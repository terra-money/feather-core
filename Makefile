#!/usr/bin/make -f

ifeq ($(OS),Windows_NT)
  $(error "Windows is not supported")
endif

FEATHER_CORE_VERSION = v0.1.0

GO := $(shell command -v go 2> /dev/null)
GOBIN = $(GOPATH)/bin
GO_VERSION := $(shell cat go.mod | grep -E 'go [0-9].[0-9]+' | cut -d ' ' -f 2)

LEDGER_ENABLED ?= true
BUILDDIR ?= $(CURDIR)/build
DOCKER := $(shell which docker)
COMMIT := $(shell git log -1 --format='%H')
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
JQ := $(shell which jq)
HTTPS_GIT := https://github.com/terra-money/feather-core.git

export GO111MODULE = on

# ensure jq is installed

ifeq ($(JQ),)
  $(error "jq" is not installed. Please install it with your package manager.)
endif

# read feather config
FEATH_CONFIG := $(CURDIR)/config/config.json

# these keys must match those in config.json
KEY_APP_NAME=app_name

# check that required keys are defined in config.json
HAS_APP_NAME := $(shell jq 'has("$(KEY_APP_NAME)")' $(FEATH_CONFIG))

ifeq ($(HAS_APP_NAME),false)
  $(error "$(FEATH_CONFIG) does not have key $(KEY_APP_NAME)")
endif

# retrieve key values, strip double quotes
FEATH_CONFIG_APP_NAME := $(patsubst "%",%,$(shell jq '.$(KEY_APP_NAME)' $(FEATH_CONFIG)))
FEATH_CONFIG_APP_BINARY_NAME := $(FEATH_CONFIG_APP_NAME)d

# process build tags
build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
empty = $(whitespace) $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(empty),$(comma),$(build_tags))

# process linker flags
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=$(FEATH_CONFIG_APP_NAME) \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=$(FEATH_CONFIG_APP_BINARY_NAME) \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(FEATHER_CORE_VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq ($(WITH_CLEVELDB),yes)
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags_comma_sep)" -ldflags '$(ldflags)' -trimpath

# The below include contains the tools and runsim targets.
include contrib/devtools/Makefile

all: install lint test

install: go.sum
	go build -o $(GOBIN)/$(FEATH_CONFIG_APP_BINARY_NAME) -mod=readonly $(BUILD_FLAGS) ./cmd/feather-cored

build: go.sum
ifeq ($(OS),Windows_NT)
	exit 1
else
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/$(FEATH_CONFIG_APP_BINARY_NAME) ./cmd/feather-cored
endif

build-contract-tests-hooks:
ifeq ($(OS),Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/contract_tests.exe ./cmd/contract_tests
else
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/contract_tests ./cmd/contract_tests
endif

########################################
### Tools & dependencies
########################################

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

draw-deps:
	@# requires brew install graphviz or apt-get install graphviz
	go install github.com/RobotsAndPencils/goviz@latest
	@goviz -i ./cmd/feather-cored -d 2 | dot -Tpng -o dependency-graph.png

clean:
	rm -rf snapcraft-local.yaml build/

distclean: clean
	rm -rf vendor/

###############################################################################
###                                Testing                                  ###
###############################################################################

test: test-unit

# For feather to use to test feather-cored correctness. E.g. make --jobs=4 test-all
test-all: test-unit test-race simulate

test-unit:
	@echo "Running unit tests..."
	@go test -mod=readonly ./...

test-race:
	@echo "Running tests with race condition detection..."
	@go test -mod=readonly -race ./...

# Generates a test coverage report, which can be used with the `go tool cover` command to view test coverage.
test-cover:
	@echo "Generating coverage profile 'coverage.out'..."
	@go test -mod=readonly -timeout 30m -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Coverage profile generated. Open in a web browser with: go tool cover -html=coverage.out"

test-benchmark:
	@go test -mod=readonly -bench=. ./...

# Convenience target for running all simulation tests.
simulate: simulate-full-app simulate-nondeterminism simulate-app-import-export

# Runs the simulation, checking invariants every operation.
simulate-full-app:
	@echo "Running full application simulation..."
	@$(GO) test ./app -run=TestFullAppSimulation \
		-mod=readonly -Enabled=true -NumBlocks=100 -BlockSize=50 -Commit=true -Period=0 -v -timeout=24h

# Runs the same simulation multiple times, verifying that the resulting app hash is the same each time.
simulate-nondeterminism:
	@echo "Running non-determinism simulation..."
	@$(GO) test ./app -run=TestAppStateDeterminism \
		-mod=readonly -Enabled=true -NumBlocks=100 -BlockSize=50 -Commit=true -Period=0 -v -timeout=24h

# Exports and imports genesis state, verifying that no data is lost in the process.
simulate-app-import-export:
	@echo "Running genesis export/import simulation..."
	@$(GO) test ./app -run=TestAppImportExport \
		-mod=readonly -Enabled=true -NumBlocks=100 -BlockSize=50 -Commit=true -Period=0 -v -timeout=24h

integration-test: clean-integration-test-data install
	@echo "Initializing both blockchains..."
	./scripts/tests/start.sh
	@echo "Create relayer..."
	./scripts/tests/relayer/rly-init.sh
	@echo "Transfer coin from chain test-1 to test-2..."
	./scripts/tests/feather/transfer.sh
	@echo "Validate the execution of ibc-hooks requests..."
	./scripts/tests/ibc-hooks/increment.sh
	@echo "Validate the creation of alliance throught feather..."
	./scripts/tests/feather/validate-alliance.sh
	@echo "Stopping feather-cored and removing previous data"
	-@rm -rf ./.test-data
	-@killall feather-cored 2>/dev/null
	-@killall rly 2>/dev/null

clean-integration-test-data:
	@echo "Stopping feather-cored and removing previous data"
	-@rm -rf ./.test-data
	-@killall feather-cored 2>/dev/null
	-@killall rly 2>/dev/null

.PHONY: integration-test clean-integration-test-data

###############################################################################
###                                Linting                                  ###
###############################################################################

format-tools:
	go install mvdan.cc/gofumpt@v0.5.0
	go install github.com/client9/misspell/cmd/misspell@v0.3.4

lint: format-tools
	golangci-lint run --tests=false
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "*_test.go" | xargs gofumpt -d -s

format: format-tools
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofumpt -w 
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs goimports -w -local github.com/CosmWasm/wasmd


###############################################################################
###                                Protobuf                                 ###
###############################################################################
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace ghcr.io/cosmos/proto-builder

proto-all: proto-format proto-lint proto-gen format

proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh

proto-format:
	@echo "Formatting Protobuf files"
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	sh ./scripts/protoc-swagger-gen.sh

proto-lint:
	@echo "Lint Protobuf files"
	@$(protoImage) buf lint --error-format=json

proto-check-breaking:
	@echo "Check Protobuf breaking changes"
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main

.PHONY: all install install-debug \
	go-mod-cache draw-deps clean build format \
	test test-all test-build test-cover test-unit \
	test-race simulate test-sim-import-export \