#!/usr/bin/env bash

hermes --config scripts/relayer_hermes/config_osmosis_gaia.toml keys delete --chain testing --all
hermes --config scripts/relayer_hermes/config_osmosis_gaia.toml keys add --chain testing --mnemonic-file scripts/relayer_hermes/alice.json

hermes --config scripts/relayer_hermes/config_osmosis_gaia.toml keys delete --chain gaiad-t1 --all
hermes --config scripts/relayer_hermes/config_osmosis_gaia.toml keys add --chain gaiad-t1 --mnemonic-file scripts/relayer_hermes/gnad.json

hermes --config scripts/relayer_hermes/config_osmosis_gaia.toml start
