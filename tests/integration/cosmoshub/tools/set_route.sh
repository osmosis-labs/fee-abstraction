#!/bin/sh

DENOM=ibc/62DDD1848B6F9182A658B6E88D7625F71D863E678ED043660692BD307C9217A0
POOL_ID=408
SWAPROUTER=osmo1j48ncj9wkzs3pnkux96ct6peg7rznnt4jx6ysdcs0283ysxj2ztqtr602y
NODE=https://osmosis-testnet-rpc.polkachu.com:443
CHAIN_ID=osmo-test-5

SET_ROUTE_1='{"set_route":{"input_denom":"'$DENOM'","output_denom":"uosmo","pool_route":[{"pool_id":"'$POOL_ID'","token_out_denom":"uosmo"}]}}'

osmosisd tx wasm execute $SWAPROUTER $SET_ROUTE_1 --node $NODE --chain-id $CHAIN_ID --gas-prices 0.1uosmo --gas auto --gas-adjustment 1.3 --from swap -y

SET_ROUTE_2='{"set_route":{"input_denom":"uosmo","output_denom":"'$DENOM'","pool_route":[{"pool_id":"'$POOL_ID'","token_out_denom":"'$DENOM'"}]}}'

sleep 5

osmosisd tx wasm execute $SWAPROUTER $SET_ROUTE_2 --node $NODE --chain-id $CHAIN_ID --gas-prices 0.1uosmo --gas auto --gas-adjustment 1.3 --from swap -y
