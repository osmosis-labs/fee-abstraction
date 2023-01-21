#!/usr/bin/env bash

osmosisd tx wasm store scripts/fee_abstraction.wasm --keyring-backend=test --home=$HOME/.osmosisd/validator1 --from validator1 --chain-id testing --gas 10000000 --fees 25000stake --yes

sleep 2

ID=13

INIT='{"packet_lifetime":100}'
osmosisd tx wasm instantiate $ID "$INIT" --keyring-backend=test --home=$HOME/.osmosisd/validator1 --from validator1 --chain-id testing --label "test" --no-admin --yes

CONTRACT=$(osmosisd query wasm list-contract-by-code $ID --output json | jq -r '.contracts[-1]')

query_params='{"query_stargate_twap":{"pool_id":1,"token_in_denom":"uosmo","token_out_denom":"uatom","with_swap_fee":false}}'
osmosisd query wasm contract-state smart $CONTRACT "$query_params"

echo "feeabs contract: "
echo $CONTRACT

osmosisd q wasm contract-state smart osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9 '{"get_route":{"input_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","output_denom":"uosmo"}}'