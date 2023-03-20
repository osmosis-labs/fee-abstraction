#!/usr/bin/env bash

hermes --config scripts/relayer_hermes/config_feeabs_gaia.toml keys delete --chain gaiad-t1 --all
hermes --config scripts/relayer_hermes/config_feeabs_gaia.toml keys add --chain gaiad-t1 --mnemonic-file scripts/relayer_hermes/gnad.json

hermes --config scripts/relayer_hermes/config_feeabs_gaia.toml keys delete --chain feeappd-t1 --all
hermes --config scripts/relayer_hermes/config_feeabs_gaia.toml keys add --chain feeappd-t1 --mnemonic-file scripts/relayer_hermes/bob.json

hermes --config scripts/relayer_hermes/config_feeabs_gaia.toml start
