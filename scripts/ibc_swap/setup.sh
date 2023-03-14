#!/usr/bin/env bash

CHANNEL_ID="channel-0"
export VALIDATOR=$(osmosisd keys show validator1 -a --keyring-backend test )
echo $VALIDATOR=$(osmosisd keys show validator1 -a --keyring-backend test )

export OWNER=$(osmosisd keys show deployer -a --keyring-backend test )
echo $OWNER=$(osmosisd keys show deployer -a --keyring-backend test )

hermes --config scripts/relayer_hermes/config.toml create channel --a-chain testing --b-chain feeappd-t1 --a-port transfer --b-port transfer --new-client-connection --yes
#hermes --config scripts/relayer_hermes/config.toml create channel --a-chain testing --b-chain feeappd-t1 --a-port icqhost --b-port feeabs --new-client-connection --yes

feeappd tx ibc-transfer transfer transfer $CHANNEL_ID "$VALIDATOR" 1000000000000stake --from feeacc --keyring-backend test --chain-id feeappd-t1 --yes --fees 5000stake
sleep 20 
echo $(osmosisd q bank balances "$VALIDATOR")

DENOM=$(osmosisd q bank balances "$VALIDATOR" -o json | jq -r '.balances[] | select(.denom | contains("ibc")) | .denom')
echo ============DENOM==============
echo $DENOM
echo ============DENOM==============

cat > sample_pool.json <<EOF
{
        "weights": "1${DENOM},1uosmo",
        "initial-deposit": "500000000000${DENOM},100000000000uosmo",
        "swap-fee": "0.01",
        "exit-fee": "0.01",
        "future-governor": "168h"
}
EOF

osmosisd tx gamm create-pool --pool-file sample_pool.json --from validator1 --keyring-backend=test --home=$HOME/.osmosisd --chain-id testing --yes --fees 5000stake
sleep 5
# get the pool id
POOL_ID=$(osmosisd query gamm pools -o json | jq -r '.pools[-1].id')

# Store the swaprouter contract
osmosisd tx wasm store scripts/bytecode/swaprouter.wasm --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --gas 10000000 --fees 25000stake --yes
# get the code id
sleep 5
SWAPROUTER_CODE_ID=$(osmosisd query wasm list-code -o json | jq -r '.code_infos[-1].code_id')
# Instantiate the swaprouter contract
INIT_SWAPROUTER='{"owner":"'$OWNER'"}'
osmosisd tx wasm instantiate $SWAPROUTER_CODE_ID "$INIT_SWAPROUTER" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --label "test" --no-admin --yes --fees 5000stake
sleep 5
SWAPROUTER_ADDRESS=$(osmosisd query wasm list-contract-by-code "$SWAPROUTER_CODE_ID" -o json | jq -r '.contracts | [last][0]')
echo $SWAPROUTER_ADDRESS

# Configure the swaprouter
#CONFIG_SWAPROUTER='{"set_route":{"input_denom":"uosmo","output_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","pool_route":[{"pool_id":"1","token_out_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"}]}}'

CONFIG_SWAPROUTER='{"set_route":{"input_denom":"uosmo","output_denom":"'$DENOM'","pool_route":[{"pool_id":"1","token_out_denom":"'$DENOM'"}]}}'

echo ==========================
echo $CONFIG_SWAPROUTER
echo ==========================

osmosisd tx wasm execute $SWAPROUTER_ADDRESS "$CONFIG_SWAPROUTER" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing -y --fees 5000stake
sleep 5

# Store the crosschainswap contract
osmosisd tx wasm store scripts/bytecode/crosschain_swaps.wasm --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --gas 10000000 --fees 25000stake --yes
# get the code id
sleep 10
CROSSCHAIN_SWAPS_CODE_ID=$(osmosisd query wasm list-code -o json | jq -r '.code_infos[-1].code_id')
# Instantiate the crosschainswap contract
INIT_CROSSCHAIN_SWAPS='{"swap_contract":"'$SWAPROUTER_ADDRESS'","channels":[["feeabs","'$CHANNEL_ID'"]]}'
echo =========INIT_CROSSCHAIN_SWAPS============
echo $INIT_CROSSCHAIN_SWAPS
echo ========INIT_CROSSCHAIN_SWAPS=============

osmosisd tx wasm instantiate $CROSSCHAIN_SWAPS_CODE_ID "$INIT_CROSSCHAIN_SWAPS" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --label "test" --no-admin --yes --fees 5000stake
sleep 5
CROSSCHAIN_SWAPS_ADDRESS=$(osmosisd query wasm list-contract-by-code "$CROSSCHAIN_SWAPS_CODE_ID" -o json | jq -r '.contracts | [last][0]')
echo $CROSSCHAIN_SWAPS_ADDRESS
#feeacc=$(feeappd keys show feeacc --keyring-backend test -a)
#balances=$(feeappd query bank balances "$feeacc" -o json | jq '.balances')


osmosisd tx ibc-transfer transfer transfer $CHANNEL_ID feeabs1efd63aw40lxf3n4mhf7dzhjkr453axurwrhrrw 100000000000uosmo --from validator1 --keyring-backend test --chain-id testing --yes --fees 5000stake
#feeappd query bank balances feeabs1efd63aw40lxf3n4mhf7dzhjkr453axurwrhrrw
#MEMO='{"wasm":{"contract":"'$CROSSCHAIN_SWAPS_ADDRESS'","msg":{"osmosis_swap":{"input_coin":{"denom":"uosmo","amount":"25000000"},"output_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","slippage":{"twap":{"slippage_percentage":"20","window_seconds":10}},"receiver":"feeabs1efd63aw40lxf3n4mhf7dzhjkr453axurwrhrrw","on_failed_delivery":"do_nothing"}}}}'
#echo $MEMO

#feeappd tx ibc-transfer transfer transfer channel-0 $CROSSCHAIN_SWAPS_ADDRESS 25000000ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518 --from feeacc --keyring-backend test --chain-id feeappd-t1 -y   --memo "$MEMO"

#sleep 20  # wait for the roundtrip

#new_balances=$(feeappd query bank balances "$feeacc" -o json | jq '.balances')
#echo "old balances: $balances, new balances: $new_balances"

