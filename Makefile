#!/usr/bin/make -f

SIMAPP = ./app
LEDGER_ENABLED ?= true
BINDIR ?= $(GOPATH)/bin
BUILDDIR ?= $(CURDIR)/build
DOCKER := $(shell which docker)
COMMIT := $(shell git log -1 --format='%H')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
GO_VERSION := $(shell cat go.mod | grep -E 'go [0-9].[0-9]+' | cut -d ' ' -f 2)
JQ := $(shell which jq)
HTTPS_GIT := https://github.com/terra-money/feather-core.git

export GO111MODULE = on

# ensure jq is installed

ifeq ($(JQ),)
  $(error "jq" is not installed. Please install it with your package manager.)
endif

# read feather config

FEATH_CONFIG := $(CURDIR)/config/mainnet/config.json

# these keys must match config/mainnet/config.json
KEY_APP_NAME=app_name
KEY_BOND_DENOM=bond_denom
KEY_APP_BINARY_NAME=app_binary_name
KEY_ACC_ADDR_PREFIX=account_address_prefix
KEY_ACC_PUBKEY_PREFIX=account_pubkey_prefix
KEY_VALIDATOR_ADDRESS_PREFIX=validator_address_prefix
KEY_VALIDATOR_PUBKEY_PREFIX=validator_pubkey_prefix
KEY_CONS_NODE_ADDR_PREFIX=consensus_node_address_prefix
KEY_CONS_NODE_PUBKEY_PREFIX=consensus_node_pubkey_prefix

# check that required keys are defined in config.json
HAS_APP_NAME := $(shell jq 'has("$(KEY_APP_NAME)")' $(FEATH_CONFIG))
HAS_BOND_DENOM := $(shell jq 'has("$(KEY_BOND_DENOM)")' $(FEATH_CONFIG))
HAS_APP_BINARY_NAME := $(shell jq 'has("$(KEY_APP_BINARY_NAME)")' $(FEATH_CONFIG))
HAS_ACC_ADDR_PREFIX := $(shell jq 'has("$(KEY_ACC_ADDR_PREFIX)")' $(FEATH_CONFIG))
HAS_ACC_PUBKEY_PREFIX := $(shell jq 'has("$(KEY_ACC_PUBKEY_PREFIX)")' $(FEATH_CONFIG))
HAS_VALIDATOR_ADDRESS_PREFIX := $(shell jq 'has("$(KEY_VALIDATOR_ADDRESS_PREFIX)")' $(FEATH_CONFIG))
HAS_VALIDATOR_PUBKEY_PREFIX := $(shell jq 'has("$(KEY_VALIDATOR_PUBKEY_PREFIX)")' $(FEATH_CONFIG))
HAS_CONS_NODE_ADDR_PREFIX := $(shell jq 'has("$(KEY_CONS_NODE_ADDR_PREFIX)")' $(FEATH_CONFIG))
HAS_CONS_NODE_PUBKEY_PREFIX := $(shell jq 'has("$(KEY_CONS_NODE_PUBKEY_PREFIX)")' $(FEATH_CONFIG))

ifeq ($(HAS_APP_NAME),false)
  $(error "$(FEATH_CONFIG) does not have key $(KEY_APP_NAME)")
endif
ifeq ($(HAS_BOND_DENOM),false)
  $(error "$(FEATH_CONFIG) does not have key $(KEY_BOND_DENOM)")
endif
ifeq ($(HAS_APP_BINARY_NAME),false)
  $(error "$(FEATH_CONFIG) does not have key $(KEY_APP_BINARY_NAME)")
endif
ifeq ($(HAS_ACC_ADDR_PREFIX),false)
  $(error "$(FEATH_CONFIG) does not have key $(KEY_ACC_ADDR_PREFIX)")
endif
ifeq ($(HAS_ACC_PUBKEY_PREFIX),false)
  $(error "$(FEATH_CONFIG) does not have key $(KEY_ACC_PUBKEY_PREFIX)")
endif
ifeq ($(HAS_VALIDATOR_ADDRESS_PREFIX),false)
  $(error "$(FEATH_CONFIG) does not have key $(KEY_VALIDATOR_ADDRESS_PREFIX)")
endif
ifeq ($(HAS_VALIDATOR_PUBKEY_PREFIX),false)
  $(error "$(FEATH_CONFIG) does not have key $(KEY_VALIDATOR_PUBKEY_PREFIX)")
endif
ifeq ($(HAS_CONS_NODE_ADDR_PREFIX),false)
  $(error "$(FEATH_CONFIG) does not have key $(KEY_CONS_NODE_ADDR_PREFIX)")
endif
ifeq ($(HAS_CONS_NODE_PUBKEY_PREFIX),false)
  $(error "$(FEATH_CONFIG) does not have key $(KEY_CONS_NODE_PUBKEY_PREFIX)")
endif

# retrieve key values, strip double quotes
FEATH_CONFIG_APP_NAME := $(patsubst "%",%,$(shell jq '.$(KEY_APP_NAME)' $(FEATH_CONFIG)))
FEATH_CONFIG_BOND_DENOM := $(patsubst "%",%,$(shell jq '.$(KEY_BOND_DENOM)' $(FEATH_CONFIG)))
FEATH_CONFIG_APP_BINARY_NAME := $(patsubst "%",%,$(shell jq '.$(KEY_APP_BINARY_NAME)' $(FEATH_CONFIG)))
FEATH_CONFIG_ACC_ADDR_PREFIX := $(patsubst "%",%,$(shell jq '.$(KEY_ACC_ADDR_PREFIX)' $(FEATH_CONFIG)))
FEATH_CONFIG_ACC_PUBKEY_PREFIX := $(patsubst "%",%,$(shell jq '.$(KEY_ACC_PUBKEY_PREFIX)' $(FEATH_CONFIG)))
FEATH_CONFIG_VALIDATOR_ADDRESS_PREFIX := $(patsubst "%",%,$(shell jq '.$(KEY_VALIDATOR_ADDRESS_PREFIX)' $(FEATH_CONFIG)))
FEATH_CONFIG_VALIDATOR_PUBKEY_PREFIX := $(patsubst "%",%,$(shell jq '.$(KEY_VALIDATOR_PUBKEY_PREFIX)' $(FEATH_CONFIG)))
FEATH_CONFIG_CONS_NODE_ADDR_PREFIX := $(patsubst "%",%,$(shell jq '.$(KEY_CONS_NODE_ADDR_PREFIX)' $(FEATH_CONFIG)))
FEATH_CONFIG_CONS_NODE_PUBKEY_PREFIX := $(patsubst "%",%,$(shell jq '.$(KEY_CONS_NODE_PUBKEY_PREFIX)' $(FEATH_CONFIG)))

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
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
		  -X github.com/terra-money/feather-core/app.AppName=$(FEATH_CONFIG_APP_NAME) \
		  -X github.com/terra-money/feather-core/app.AccountAddressPrefix=$(FEATH_CONFIG_ACC_ADDR_PREFIX) \
		  -X github.com/terra-money/feather-core/app.AccountPubKeyPrefix=$(FEATH_CONFIG_ACC_PUBKEY_PREFIX) \
		  -X github.com/terra-money/feather-core/app.ValidatorAddressPrefix=$(FEATH_CONFIG_VALIDATOR_ADDRESS_PREFIX) \
		  -X github.com/terra-money/feather-core/app.ValidatorPubKeyPrefix=$(FEATH_CONFIG_VALIDATOR_PUBKEY_PREFIX) \
		  -X github.com/terra-money/feather-core/app.ConsensusNodeAddressPrefix=$(FEATH_CONFIG_CONS_NODE_ADDR_PREFIX) \
		  -X github.com/terra-money/feather-core/app.ConsensusNodePubKeyPrefix=$(FEATH_CONFIG_ACC_PUBKEY_PREFIX) \
		  -X github.com/terra-money/feather-core/app.BondDenom=$(FEATH_CONFIG_BOND_DENOM)

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
	go build -o $(BINDIR)/$(FEATH_CONFIG_APP_BINARY_NAME) -mod=readonly $(BUILD_FLAGS) ./cmd/feather-core

build: go.sum
ifeq ($(OS),Windows_NT)
	exit 1
else
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/feather-core ./cmd/feather-core
endif

build-mantlemint: go.sum
ifeq ($(OS),Windows_NT)
	exit 1
else
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/mantlemint ./mantlemint/sync.go
endif

build-contract-tests-hooks:
ifeq ($(OS),Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/contract_tests.exe ./cmd/contract_tests
else
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/contract_tests ./cmd/contract_tests
endif

build-reproducible: build-reproducible-amd64 build-reproducible-arm64

build-reproducible-amd64: go.sum $(BUILDDIR)/
	$(DOCKER) buildx create --name feather-core-builder || true
	$(DOCKER) buildx use feather-core-builder
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg RUNNER_IMAGE=alpine:3.16 \
		--platform linux/amd64 \
		-t feather-core:local-amd64 \
		--load \
		-f Dockerfile .
	$(DOCKER) rm -f feather-core-binary || true
	$(DOCKER) create -ti --name feather-core-binary feather-core:local-amd64
	$(DOCKER) cp feather-core-binary:/usr/local/bin/feather-core $(BUILDDIR)/feather-core-linux-amd64
	$(DOCKER) rm -f feather-core-binary

build-reproducible-arm64: go.sum $(BUILDDIR)/
	$(DOCKER) buildx create --name feather-core-builder  || true
	$(DOCKER) buildx use feather-core-builder 
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg RUNNER_IMAGE=alpine:3.16 \
		--platform linux/arm64 \
		-t feather-core:local-arm64 \
		--load \
		-f Dockerfile .
	$(DOCKER) rm -f feather-core-binary || true
	$(DOCKER) create -ti --name feather-core-binary feather-core:local-arm64
	$(DOCKER) cp feather-core-binary:/usr/local/bin/feather-core $(BUILDDIR)/feather-core-linux-arm64
	$(DOCKER) rm -f feather-core-binary

########################################
### Tools & dependencies

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

draw-deps:
	@# requires brew install graphviz or apt-get install graphviz
	go install github.com/RobotsAndPencils/goviz@latest
	@goviz -i ./cmd/feather-core -d 2 | dot -Tpng -o dependency-graph.png

clean:
	rm -rf snapcraft-local.yaml build/

distclean: clean
	rm -rf vendor/

###############################################################################
###                                Testing                                  ###
###############################################################################

test: test-unit
test-all: test-race test-cover

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./...

test-race:
	@VERSION=$(VERSION) go test -mod=readonly -race -tags='ledger test_ledger_mock' ./...

test-cover:
	@go test -mod=readonly -timeout 30m -race -coverprofile=coverage.txt -covermode=atomic -tags='ledger test_ledger_mock' ./...

benchmark:
	@go test -mod=readonly -bench=. ./...

test-sim-import-export: runsim
	@echo "Running application import/export simulation. This may take several minutes..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 5 TestAppImportExport

test-sim-multi-seed-short: runsim
	@echo "Running short multi-seed application simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 10 TestFullAppSimulation
	
simulate:
	@go test -v -run=TestFullAppSimulation ./app -NumBlocks 200 -BlockSize 50 -Commit -Enabled -Period 1

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