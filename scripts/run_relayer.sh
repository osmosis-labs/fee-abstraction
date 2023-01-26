#!/usr/bin/env bash

hermes --config scripts/relayer_hermes/config.toml keys delete --chain testing --all
hermes --config scripts/relayer_hermes/config.toml keys add --chain testing --mnemonic-file scripts/relayer_hermes/alice.json

hermes --config scripts/relayer_hermes/config.toml keys delete --chain feeappd-t1 --all
hermes --config scripts/relayer_hermes/config.toml keys add --chain feeappd-t1 --mnemonic-file scripts/relayer_hermes/bob.json

hermes --config scripts/relayer_hermes/config.toml start
hermes --config scripts/relayer_hermes/config.toml create channel --a-chain feeappd-t1 --b-chain testing --a-port transfer --b-port transfer  --new-client-connection
hermes --config scripts/relayer_hermes/config.toml create channel --a-chain feeappd-t1 --b-chain testing --a-port feeabs --b-port wasm.osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9 --channel-version simple-ica-v1 --new-client-connection

osmosisd tx gamm create-pool --pool-file scripts/pool.json --from osmo1hj5fveer5cjtn4wd6wstzugjfdxzl0xpwhpz63 --yes --chain-id testing --keyring-backend test

osmo1hj5fveer5cjtn4wd6wstzugjfdxzl0xpwhpz63 : 100005600000 ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878
feeabs1efd63aw40lxf3n4mhf7dzhjkr453axurwrhrrw : 10000000 ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518

feeappd tx bank send feeabs1efd63aw40lxf3n4mhf7dzhjkr453axurwrhrrw feeabs1hq6049htg8dh9swl5cw6uqqqcasxttdv4ynj83 10000000ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518 --keyring-backend test --chain-id
 feeappd-t1
 
params :
- OsmosisIBCDenom: ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518
- NativeIBCDenom: ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878
- TransferChannel: channel-0
- PoolID: 1
- QueryChannel: channel-1
- SwapContract: osmo1unyuj8qnmygvzuex3dwmg9yzt9alhvyeat0uu0jedg2wj33efl5q0keax0
- QueryContractAddress: osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9