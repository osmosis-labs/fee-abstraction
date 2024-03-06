#!/bin/sh

DENOM=ibc/80C64E7EB7E8B6705FC9C1D9C486EB6278823068D9224915B6A5DABDF03FB2D5
POOL_ID=409
SWAPROUTER=osmo1j48ncj9wkzs3pnkux96ct6peg7rznnt4jx6ysdcs0283ysxj2ztqtr602y
NODE=https://osmosis-testnet-rpc.polkachu.com:443
CHAIN_ID=osmo-test-5

SET_ROUTE_1='{"set_route":{"input_denom":"'$DENOM'","output_denom":"uosmo","pool_route":[{"pool_id":"'$POOL_ID'","token_out_denom":"uosmo"}]}}'

osmosisd tx wasm execute $SWAPROUTER $SET_ROUTE_1 --node $NODE --chain-id $CHAIN_ID --gas-prices 0.1uosmo --gas auto --gas-adjustment 1.3 --from swap -y

SET_ROUTE_2='{"set_route":{"input_denom":"uosmo","output_denom":"'$DENOM'","pool_route":[{"pool_id":"'$POOL_ID'","token_out_denom":"'$DENOM'"}]}}'

sleep 5

osmosisd tx wasm execute $SWAPROUTER $SET_ROUTE_2 --node $NODE --chain-id $CHAIN_ID --gas-prices 0.1uosmo --gas auto --gas-adjustment 1.3 --from swap -y
