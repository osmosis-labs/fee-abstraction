#!/usr/bin/env bash

ROOT=$(pwd)

# underscore so that go tool will not take gocache into account
mkdir -p _build/gocache
export GOMODCACHE=$ROOT/_build/gocache

# install gaid binary
if ! command -v _build/binary/gaid &> /dev/null
then
    echo "Building gaiad..."
    cd deps/gaia
    GOBIN="$ROOT/_build/binary" go install -mod=readonly ./...
    cd ../..
fi

# install osmosisd binary
if ! command -v _build/binary/osmosisd &> /dev/null
then
    echo "Building osmosisd..."
    cd deps/osmosis
    GOBIN="$ROOT/_build/binary" go install -mod=readonly ./...
    cd ../..
fi

# install relayer binary
if ! command -v _build/binary/relayer &> /dev/null
then
    echo "Building relayer..."
    cd deps/relayer
    GOBIN="$ROOT/_build/binary" go install -mod=readonly ./...
    cd ../..
fi

# install stargaze binary
if ! command -v _build/binary/starsd &> /dev/null
then
    echo "Building starsd..."
    cd deps/stargaze
    GOBIN="$ROOT/_build/binary" go install -mod=readonly ./...
    cd ../..
fi