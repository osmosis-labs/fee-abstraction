#!/usr/bin/env bash

osmosisd tx wasm store scripts/swaprouter.wasm --keyring-backend=test --from validator1 --chain-id testing --gas 10000000 --fees 25000stake --yes

sleep 2

ID=3

INIT='{"packet_lifetime":100}'
osmosisd tx wasm instantiate 5 '{"swap_contract": "osmo1466nf3zuxpya8q9emxukd7vftaf6h4psr0a07srl5zw74zh84yjqkk0zfx", "channels": [["cosmos", "channel-3"]]}' --keyring-backend=test --from validator1 --chain-id testing --label "test" --no-admin --yes
osmosisd tx wasm instantiate 6 '{"owner": "osmo18ww3tl6c9vj88wlpsvfuuxmjzxy0grn9lzgudk"}' --keyring-backend=test --from validator1 --chain-id testing --label "test" --no-admin --yes

CONTRACT=$(osmosisd query wasm list-contract-by-code 1 --output json | jq -r '.contracts[-1]')

query_params='{"query_stargate_twap":{"pool_id":1,"token_in_denom":"uosmo","token_out_denom":"uatom","with_swap_fee":false}}'
osmosisd query wasm contract-state smart osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9 "$query_params"

osmosisd tx gamm create-pool --pool-file scripts/pool.json --from validator1 --yes --chain-id testing

echo "feeabs contract: "
echo $CONTRACT

MSG_INIT_SWAP = '{"owner": "osmo18ww3tl6c9vj88wlpsvfuuxmjzxy0grn9lzgudk"}'
MSG_CROSS= '{"swap_contract": "osmo1wkwy0xh89ksdgj9hr347dyd2dw7zesmtrue6kfzyml4vdtz6e5wsfdyyaj", "channels": [["osmo", "channel-3"]]}'
MSG_SWAPROUTER= '{"set_route":{"input_denom":"ibc/4D74FBE09BED153381B75FF0D0B030A839E68AE17761F3945A8AF5671B915928","output_denom":"uosmo","pool_route":[{"pool_id":"2","token_out_denom":"uosmo"}]}}'
MSG_SWAPROUTER= '{"set_route":{"input_denom":"uosmo","output_denom":"ibc/4D74FBE09BED153381B75FF0D0B030A839E68AE17761F3945A8AF5671B915928","pool_route":[{"pool_id":"2","token_out_denom":"uosmo"}]}}'

cross_swap_addr := osmo1eyfccmjm6732k7wp4p6gdjwhxjwsvje44j0hfx8nkgrm8fs7vqfsn92ayh
osmosisd tx wasm execute osmo1466nf3zuxpya8q9emxukd7vftaf6h4psr0a07srl5zw74zh84yjqkk0zfx '{"set_route":{"input_denom":"uosmo","output_denom":"ibc/4D74FBE09BED153381B75FF0D0B030A839E68AE17761F3945A8AF5671B915928","pool_route":[{"pool_id":"2","token_out_denom":"ibc/4D74FBE09BED153381B75FF0D0B030A839E68AE17761F3945A8AF5671B915928"}]}}' --keyring-backend=test --from validator1 --chain-id testing --yes