#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')

ifeq (,$(VERSION))
  VERSION := $(shell git describe --tags)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
TM_VERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::') 
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf:1.0.0-rc8
BUILDDIR ?= $(CURDIR)/build
HTTPS_GIT := https://github.com/osmosis-labs/fee-abstraction.git

export GO111MODULE = on

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

ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += gcc cleveldb
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=feeapp \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=feeappd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
			-X github.com/cometbft/cometbft/version.TMCoreSemVer=$(TM_VERSION)

ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq ($(LINK_STATICALLY),true)
  ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif


all: install

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/feeappd

build:
	go build $(BUILD_FLAGS) -o bin/feeappd ./cmd/feeappd

test:
	go test -race -v ./... 

docker-build-debug:
	@DOCKER_BUILDKIT=1 docker build -t feeapp:debug -f Dockerfile .

lint:
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' -not -name '*.gw.go' | xargs go run mvdan.cc/gofumpt -w .
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' -not -name '*.gw.go' | xargs go run github.com/client9/misspell/cmd/misspell -w
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' -not -name '*.gw.go' | xargs go run golang.org/x/tools/cmd/goimports -w -local github.com/osmosis-labs/fee-abstraction
.PHONY: lint test

###############################################################################
###                             Interchain test                             ###
###############################################################################

# Executes basic chain tests via interchaintest
ictest-basic:
	cd tests/interchaintest && go test -race -v -run TestStartFeeabs .

# Executes IBC tests via interchaintest
ictest-ibc:
	cd tests/interchaintest && go test -timeout=25m -race -v -run TestFeeabsGaiaIBCTransfer .

# Executes IBC tests via interchaintest
ictest-packet-forward:
	cd tests/interchaintest && go test -timeout=25m -race -v -run TestPacketForwardMiddleware .

ictest-host-zone-proposal:
	cd tests/interchaintest && go test -timeout=25m -race -v -run TestHostZoneProposal .

# Executes all tests via interchaintest after compling a local image as juno:local
ictest-all: ictest-basic ictest-ibc ictest-packet-forward

.PHONY: ictest-basic ictest-ibc ictest-packet-forward ictest-all

###############################################################################
###                                  Proto                                  ###
###############################################################################
PROTO_BUILDER_IMAGE=ghcr.io/cosmos/proto-builder:0.13.5

proto-all: proto-format proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(PROTO_BUILDER_IMAGE) sh ./scripts/protocgen.sh

proto-format:
	@echo "Formatting Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoFmt}$$"; then docker start -a $(containerProtoFmt); else docker run --name $(containerProtoFmt) -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./ -not -path "./third_party/*" -name "*.proto" -exec clang-format -i {} \; ; fi

proto-lint:
	@$(DOCKER_BUF) lint --error-format=json

proto-check-breaking:
	@$(DOCKER_BUF) breaking --against $(HTTPS_GIT)#branch=main

proto-image-build:
	@DOCKER_BUILDKIT=1 docker build -t $(protoImageName) -f ./proto/Dockerfile ./proto

proto-image-push:
	docker push $(protoImageName)

.PHONY: proto-all proto-gen proto-format proto-lint proto-check-breaking 

###############################################################################
###                                Querygen                                 ###
###############################################################################

run-querygen:
	@go run cmd/querygen/main.go
