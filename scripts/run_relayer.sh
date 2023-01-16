#!/usr/bin/env bash

hermes --config scripts/relayer_hermes/config.toml keys delete --chain testing --all
hermes --config scripts/relayer_hermes/config.toml keys add --chain testing --mnemonic-file scripts/relayer_hermes/alice.json

hermes --config scripts/relayer_hermes/config.toml keys delete --chain feeappd-t1 --all
hermes --config scripts/relayer_hermes/config.toml keys add --chain feeappd-t1 --mnemonic-file scripts/relayer_hermes/bob.json

hermes --config scripts/relayer_hermes/config.toml start

hermes --config scripts/relayer_hermes/config.toml create channel --a-chain feeappd-t1 --b-chain testing --a-port transfer --b-port transfer  --new-client-connection
hermes --config scripts/relayer_hermes/config.toml create channel --a-chain feeappd-t1 --b-chain testing --a-port feeabs --b-port wasm.osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9 --new-client-connection --channel-version "simple-ica-v1"
