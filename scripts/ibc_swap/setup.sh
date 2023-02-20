#!/usr/bin/env bash

CHANNEL_ID="channel-0"
export VALIDATOR=$(osmosisd keys show validator1 -a --keyring-backend test )
echo $VALIDATOR

hermes --config scripts/relayer_hermes/config.toml create channel --a-chain testing --b-chain feeappd-t1 --a-port transfer --b-port transfer --new-client-connection --yes

feeappd tx ibc-transfer transfer transfer $CHANNEL_ID "$VALIDATOR" 1000000000000stake --from feeacc --keyring-backend test --chain-id feeappd-t1 --yes
sleep 22 
echo $(osmosisd q bank balances "$VALIDATOR")

DENOM=$(osmosisd q bank balances "$VALIDATOR" -o json | jq -r '.balances[] | select(.denom | contains("ibc")) | .denom')

cat > sample_pool.json <<EOF
{
        "weights": "1${DENOM},1uosmo",
        "initial-deposit": "10000000000${DENOM},10000000000uosmo",
        "swap-fee": "0.01",
        "exit-fee": "0.01",
        "future-governor": "168h"
}
EOF

osmosisd tx gamm create-pool --pool-file sample_pool.json --from validator1 --keyring-backend=test --home=$HOME/.osmosisd --chain-id testing --yes
sleep 6
# get the pool id
POOL_ID=$(osmosisd query gamm pools -o json | jq -r '.pools[-1].id')

# Store the swaprouter contract
osmosisd tx wasm store scripts/bytecode/swaprouter.wasm --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --gas 10000000 --fees 25000stake --yes
# get the code id
sleep 6
SWAPROUTER_CODE_ID=$(osmosisd query wasm list-code -o json | jq -r '.code_infos[-1].code_id')
# Instantiate the swaprouter contract
INIT_SWAPROUTER='{"owner":"'$VALIDATOR'"}'
osmosisd tx wasm instantiate $SWAPROUTER_CODE_ID "$INIT_SWAPROUTER" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --label "test" --no-admin --yes 
sleep 5
SWAPROUTER_ADDRESS=$(osmosisd query wasm list-contract-by-code "$SWAPROUTER_CODE_ID" -o json | jq -r '.contracts | [last][0]')
# Configure the swaprouter
CONFIG_SWAPROUTER='{"set_route":{"input_denom":"'$DENOM'","output_denom":"uosmo","pool_route":[{"pool_id":'$POOL_ID',"token_out_denom":"uosmo"}]}}'
osmosisd tx wasm execute "$SWAPROUTER_ADDRESS" "$CONFIG_SWAPROUTER" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing -y
sleep 5

# Store the crosschainswap contract
osmosisd tx wasm store scripts/bytecode/crosschain_swaps.wasm --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --gas 10000000 --fees 25000stake --yes
# get the code id
sleep 6
CROSSCHAIN_SWAPS_CODE_ID=$(osmosisd query wasm list-code -o json | jq -r '.code_infos[-1].code_id')
# Instantiate the crosschainswap contract
INIT_CROSSCHAIN_SWAPS='{"swap_contract":"'$SWAPROUTER_ADDRESS'","channels":[["cosmos","'$CHANNEL_ID'"]]}'
osmosisd tx wasm instantiate $CROSSCHAIN_SWAPS_CODE_ID "$INIT_CROSSCHAIN_SWAPS" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --label "test" --no-admin --yes 
sleep 5
CROSSCHAIN_SWAPS_ADDRESS=$(osmosisd query wasm list-contract-by-code "$CROSSCHAIN_SWAPS_CODE_ID" -o json | jq -r '.contracts | [last][0]')

feeacc=$(feeappd keys show feeacc --keyring-backend test -a)
balances=$(feeappd query bank balances "$feeacc" -o json | jq '.balances')

MEMO='{"wasm":{"contract":"'$CROSSCHAIN_SWAPS_ADDRESS'","msg":{"osmosis_swap":{"input_coin":{"denom":"'$DENOM'","amount":"100"},"output_denom":"uosmo","slippage":{"twap":{"slippage_percentage":"20","window_seconds":10}},"receiver":"'$feeacc'","on_failed_delivery":"do_nothing"}}}}'

feeappd tx ibc-transfer transfer transfer $CHANNEL_ID $CROSSCHAIN_SWAPS_ADDRESS 100stake \
    --from feeacc --keyring-backend test --chain-id feeappd-t1 -y \
    --memo "$MEMO"

sleep 20  # wait for the roundtrip

new_balances=$(feeappd query bank balances "$feeacc" -o json | jq '.balances')
echo "old balances: $balances, new balances: $new_balances"

