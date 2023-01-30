#!/usr/bin/env bash

osmosisd tx gamm create-pool --pool-file scripts/pool.json --from validator1 --keyring-backend=test --home=$HOME/.osmosisd/validator1 --chain-id testing --yes
