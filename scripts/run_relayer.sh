#!/usr/bin/env bash

hermes --config scripts/relayer_hermes/config.toml keys delete --chain testing --all
hermes --config scripts/relayer_hermes/config.toml keys add --chain testing --mnemonic-file scripts/relayer_hermes/alice.json

hermes --config scripts/relayer_hermes/config.toml keys delete --chain feeappd-t1 --all
hermes --config scripts/relayer_hermes/config.toml keys add --chain feeappd-t1 --mnemonic-file scripts/relayer_hermes/bob.json

hermes --config scripts/relayer_hermes/config.toml start
