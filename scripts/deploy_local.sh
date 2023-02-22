#!/usr/bin/env bash

osmosisd tx wasm store scripts/bytecode/ibc_stargate.wasm --keyring-backend=test --from validator1 --chain-id testing --gas 10000000 --fees 25000stake --yes

sleep 2

ID=3

INIT='{"packet_lifetime":1000000000}'
osmosisd tx wasm instantiate $ID "$INIT" --keyring-backend=test --from validator1 --chain-id testing --label "test" --no-admin --yes

CONTRACT=$(osmosisd query wasm list-contract-by-code 3 --output json | jq -r '.contracts[-1]')
echo $CONTRACT
query_params='{"query_stargate_twap":{"pool_id":1,"token_in_denom":"uosmo","token_out_denom":"uatom","with_swap_fee":false}}'
osmosisd query wasm contract-state smart osmo1eyfccmjm6732k7wp4p6gdjwhxjwsvje44j0hfx8nkgrm8fs7vqfsn92ayh "$query_params"

echo "feeabs contract: "
echo $CONTRACT

osmosisd q wasm contract-state smart osmo17p9rzwnnfxcjp32un9ug7yhhzgtkhvl9jfksztgw5uh69wac2pgs5yczr8 '{"get_route":{"input_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","output_denom":"uosmo"}}'
hermes --config scripts/relayer_hermes/config.toml create channel --a-chain testing --b-chain feeappd-t1 --a-port wasm.osmo17p9rzwnnfxcjp32un9ug7yhhzgtkhvl9jfksztgw5uh69wac2pgs5yczr8 --b-port feeabs --new-client-connection --yes

#osmosisd tx wasm store scripts/bytecode/ibc_stargate-aarch64.wasm --keyring-backend=test --home=$HOME/.osmosisd --from osmo1hj5fveer5cjtn4wd6wstzugjfdxzl0xpwhpz63 --chain-id testing --gas 10000000 --fees 25000stake --yes
#INIT='{"packet_lifetime":100}'
#osmosisd tx wasm instantiate 1 "$INIT" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --label "test" --no-admin --yes
#hermes --config scripts/relayer_hermes/config.toml create channel --a-chain testing --b-chain feeappd-t1 --a-port wasm.osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9 --b-port feeabs --new-client-connection --yes
#osmosisd tx gamm create-pool --pool-file scripts/pool.json --from validator1 --keyring-backend=test --home=$HOME/.osmosisd --chain-id testing --yes
#feeappd tx feeabs interchain-query osmo1ekqk6ms4fqf2mfeazju4pcu3jq93lcdsfl0tah --keyring-backend test --chain-id feeappd-t1 --from feeacc
#feeappd tx feeabs query-osmosis-twap uatom 1676778631 --keyring-backend test --chain-id feeappd-t1 --from feeacc