#!/usr/bin/env bash

hermes --config scripts/relayer_hermes/config.toml keys delete --chain testing --all
hermes --config scripts/relayer_hermes/config.toml keys add --chain testing --mnemonic-file scripts/relayer_hermes/alice.json

hermes --config scripts/relayer_hermes/config.toml keys delete --chain feeappd-t1 --all
hermes --config scripts/relayer_hermes/config.toml keys add --chain feeappd-t1 --mnemonic-file scripts/relayer_hermes/bob.json

hermes --config scripts/relayer_hermes/config.toml start


hermes --config scripts/relayer_hermes/config.toml create channel --a-chain feeappd-t1 --b-chain testing --a-port transfer --b-port transfer  --new-client-connection
hermes --config scripts/relayer_hermes/config.toml create channel --a-chain feeappd-t1 --b-chain testing --a-port feeabs --b-port wasm.osmo1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrsll0sqv --new-client-connection --channel-version "simple-ica-v1"

transfer: channel-0
query: channel-1
queryaddress: osmo1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrsll0sqv
swapaddress: osmo1unyuj8qnmygvzuex3dwmg9yzt9alhvyeat0uu0jedg2wj33efl5q0keax0
ibc native denom: ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878
ibc osmosis denom: ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518
pool_id: 1