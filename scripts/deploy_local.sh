#!/usr/bin/env bash

osmosisd tx wasm store scripts/bytecode/fee_abstraction.wasm --keyring-backend=test  --from osmo1hj5fveer5cjtn4wd6wstzugjfdxzl0xpwhpz63 --chain-id testing --gas 10000000 --fees 25000stake --yes
osmosisd tx wasm store scripts/bytecode/crosschain_swaps.wasm --keyring-backend=test --from validator1 --chain-id testing --gas 10000000 --fees 25000stake --yes

sleep 2

ID=13

osmosisd tx wasm instantiate 1 '{"packet_lifetime":1000000000}' --keyring-backend=test --from osmo1hj5fveer5cjtn4wd6wstzugjfdxzl0xpwhpz63 --chain-id testing --label "test" --no-admin --yes
osmosisd tx wasm instantiate 2 '{"owner": "osmo1hj5fveer5cjtn4wd6wstzugjfdxzl0xpwhpz63"}' --keyring-backend=test --from osmo1hj5fveer5cjtn4wd6wstzugjfdxzl0xpwhpz63 --chain-id testing --label "test" --no-admin --yes
osmosisd tx wasm instantiate 3 '{"swap_contract": "osmo1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqvlx82r", "channels": [["feeabs", "channel-0"]]}' --keyring-backend=test --from validator1 --chain-id testing --label "test" --no-admin --yes

CONTRACT=$(osmosisd query wasm list-contract-by-code $ID --output json | jq -r '.contracts[-1]')
MSG_SWAPROUTER= '{"set_route":{"input_denom":"ibc/4D74FBE09BED153381B75FF0D0B030A839E68AE17761F3945A8AF5671B915928","output_denom":"uosmo","pool_route":[{"pool_id":"2","token_out_denom":"uosmo"}]}}'
MSG_SWAPROUTER= '{"set_route":{"input_denom":"uosmo","output_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","pool_route":[{"pool_id":"2","token_out_denom":"uosmo"}]}}'

query_params='{"query_stargate_twap":{"pool_id":1,"token_in_denom":"uosmo","token_out_denom":"uatom","with_swap_fee":false}}'
osmosisd query wasm contract-state smart $CONTRACT "$query_params"

echo "feeabs contract: "
echo $CONTRACT


osmosisd q wasm contract-state smart osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9 '{"get_route":{"input_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","output_denom":"uosmo"}}'
osmosisd tx wasm execute osmo1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqvlx82r '{"set_route":{"input_denom":"uosmo","output_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","pool_route":[{"pool_id":"1","token_out_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"}]}}' --keyring-backend=test --from osmo1hj5fveer5cjtn4wd6wstzugjfdxzl0xpwhpz63 --chain-id testing --yes
osmosisd tx wasm store scripts/bytecode/swaprouter.wasm --keyring-backend=test --from validator1 --chain-id testing --gas 10000000 --fees 25000stake --yes

swaprouter addr: osmo1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqvlx82r
transfer addr: osmo17p9rzwnnfxcjp32un9ug7yhhzgtkhvl9jfksztgw5uh69wac2pgs5yczr8
swap_cross addr : osmo1unyuj8qnmygvzuex3dwmg9yzt9alhvyeat0uu0jedg2wj33efl5q0keax0