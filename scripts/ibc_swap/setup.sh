#!/usr/bin/env bash

CHANNEL_ID="channel-0"
export VALIDATOR=$(osmosisd keys show validator1 -a --keyring-backend test )
echo $VALIDATOR=$(osmosisd keys show validator1 -a --keyring-backend test )

export OWNER=$(osmosisd keys show deployer -a --keyring-backend test )
echo $OWNER=$(osmosisd keys show deployer -a --keyring-backend test )

hermes --config scripts/relayer_hermes/config_feeabs_osmosis.toml create channel --a-chain testing --b-chain feeappd-t1 --a-port transfer --b-port transfer --new-client-connection --yes
hermes --config scripts/relayer_hermes/config_feeabs_osmosis.toml create channel --a-chain testing --b-chain feeappd-t1 --a-port icqhost --b-port feeabs --new-client-connection --yes
hermes --config scripts/relayer_hermes/config_feeabs_gaia.toml create channel --a-chain gaiad-t1 --b-chain feeappd-t1 --a-port transfer --b-port transfer --new-client-connection --yes
hermes --config scripts/relayer_hermes/config_osmosis_gaia.toml create channel --a-chain gaiad-t1 --b-chain testing --a-port transfer --b-port transfer --new-client-connection --yes
#feeabs - osmo: channel-0 channel-0
#feeabs - osmo: (feeabs - icqhost) channel-1 channel-1
#feeabs - gaia: channel-2 channel-0
#osmo   - gaia: channel-2 channel-1
#
#

feeappd tx ibc-transfer transfer transfer $CHANNEL_ID "$VALIDATOR" 1000000000000stake --from myaccount --keyring-backend test --chain-id feeappd-t1 --yes --fees 5000stake
gaiad tx ibc-transfer transfer transfer channel-1 "$VALIDATOR" 1000000000000uatom --from gnad --keyring-backend test --chain-id gaiad-t1 --yes --fees 5000stake
gaiad tx ibc-transfer transfer transfer channel-0 feeabs1efd63aw40lxf3n4mhf7dzhjkr453axurwrhrrw 1000000000000uatom --from gnad --keyring-backend test --chain-id gaiad-t1 --yes --fees 5000stake
sleep 20 
echo $(osmosisd q bank balances "$VALIDATOR")

DENOM=$(osmosisd q bank balances "$VALIDATOR" -o json | jq -r '.balances[] | select(.denom | contains("ibc")) | .denom')
echo ============DENOM==============
echo $DENOM
echo ============DENOM==============

cat > sample_pool.json <<EOF
{
        "weights": "1ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E,1ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878",
        "initial-deposit": "500000000000ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E,1000000000000ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878",
        "swap-fee": "0.01",
        "exit-fee": "0",
        "future-governor": "168h"
}
EOF

osmosisd tx gamm create-pool --pool-file sample_pool.json --from validator1 --keyring-backend=test --home=$HOME/.osmosisd --chain-id testing --yes --fees 5000stake --gas 400000
sleep 5
# get the pool id
POOL_ID=$(osmosisd query gamm pools -o json | jq -r '.pools[-1].id')

#store the SetupCrosschainRegistry
osmosisd tx wasm store scripts/bytecode/crosschain_registry.wasm --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --gas 10000000 --fees 25000stake --yes
# instantiate
INIT_SWAPREGISTRY='{"owner":"'$OWNER'"}'
osmosisd tx wasm instantiate 1 "$INIT_SWAPREGISTRY" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --label "test" --no-admin --yes --fees 5000stake
SWAPREGISTRY_ADDRESS=osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9
# execute
EXE_MSG='{"modify_chain_channel_links": {"operations": [{"operation": "set","source_chain": "feeappd-t1","destination_chain": "osmosis","channel_id": "channel-0"},{"operation": "set","source_chain": "osmosis","destination_chain": "feeappd-t1","channel_id": "channel-0"},{"operation": "set","source_chain": "feeappd-t1","destination_chain": "gaiad-t1","channel_id": "channel-2"},{"operation": "set","source_chain": "gaiad-t1","destination_chain": "feeappd-t1","channel_id": "channel-0"},{"operation": "set","source_chain": "osmosis","destination_chain": "gaiad-t1","channel_id": "channel-2"},{"operation": "set","source_chain": "gaiad-t1","destination_chain": "osmosis","channel_id": "channel-1"}]}}'
osmosisd tx wasm execute osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9 "$EXE_MSG" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --yes --fees 5000stake
PREFIX='{"modify_bech32_prefixes": {"operations": [{"operation": "set", "chain_name": "feeappd-t1", "prefix": "feeabs"},{"operation": "set", "chain_name": "osmosis", "prefix": "osmo"},{"operation": "set", "chain_name": "gaiad-t1", "prefix": "cosmos"}]}}'
osmosisd tx wasm execute osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9 "$PREFIX" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --yes --fees 5000stake

#osmosisd q wasm contract-state smart osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9 '{"get_destination_chain_from_source_chain_via_channel": {"on_chain": "osmosis", "via_channel": "channel-0"}}'

# Store the swaprouter contract
osmosisd tx wasm store scripts/bytecode/swaprouter.wasm --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --gas 10000000 --fees 25000stake --yes
# get the code id
sleep 5
SWAPROUTER_CODE_ID=$(osmosisd query wasm list-code -o json | jq -r '.code_infos[-1].code_id')
# Instantiate the swaprouter contract
INIT_SWAPROUTER='{"owner":"'$OWNER'"}'
osmosisd tx wasm instantiate 2 "$INIT_SWAPROUTER" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --label "test" --no-admin --yes --fees 5000stake
sleep 5
SWAPROUTER_ADDRESS=osmo1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqvlx82r
echo $SWAPROUTER_ADDRESS

# Configure the swaprouter
#CONFIG_SWAPROUTER='{"set_route":{"input_denom":"uosmo","output_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","pool_route":[{"pool_id":"1","token_out_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"}]}}'

CONFIG_SWAPROUTER='{"set_route":{"input_denom":"ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E","output_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","pool_route":[{"pool_id":"1","token_out_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"}]}}'

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
INIT_CROSSCHAIN_SWAPS='{"swap_contract":"'$SWAPROUTER_ADDRESS'","governor": "'$OWNER'"}'
echo =========INIT_CROSSCHAIN_SWAPS============
echo $INIT_CROSSCHAIN_SWAPS
echo ========INIT_CROSSCHAIN_SWAPS=============

osmosisd tx wasm instantiate 3 "$INIT_CROSSCHAIN_SWAPS" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --label "test" --no-admin --yes --fees 5000stake
sleep 5
CROSSCHAIN_SWAPS_ADDRESS=osmo17p9rzwnnfxcjp32un9ug7yhhzgtkhvl9jfksztgw5uh69wac2pgs5yczr8
echo $CROSSCHAIN_SWAPS_ADDRESS
#feeacc=$(feeappd keys show feeacc --keyring-backend test -a)
#balances=$(feeappd query bank balances "$feeacc" -o json | jq '.balances')

osmosisd tx ibc-transfer transfer transfer $CHANNEL_ID feeabs1efd63aw40lxf3n4mhf7dzhjkr453axurwrhrrw 100000000000uosmo --from validator1 --keyring-backend test --chain-id testing --yes --fees 5000stake
#feeappd query bank balances feeabs1efd63aw40lxf3n4mhf7dzhjkr453axurwrhrrw
#feeappd tx feeabs fund 500000000ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E --from myaccount --keyring-backend test --chain-id feeappd-t1 -y
#MEMO='{"wasm":{"contract":"'$CROSSCHAIN_SWAPS_ADDRESS'","msg":{"osmosis_swap":{"output_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","slippage":{"twap":{"slippage_percentage":"20","window_seconds":10}},"receiver":"feeappd-t1/feeabs1efd63aw40lxf3n4mhf7dzhjkr453axurwrhrrw","on_failed_delivery":"do_nothing", "next_memo":{}}}}}'
#echo $MEMO

#feeappd tx ibc-transfer transfer transfer channel-0 $CROSSCHAIN_SWAPS_ADDRESS 250000000ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E --from myaccount --keyring-backend test --chain-id feeappd-t1 -y   --memo "$MEMO"

#feeappd tx ibc-transfer transfer transfer channel-0 $CROSSCHAIN_SWAPS_ADDRESS 250000000ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E --from myaccount --keyring-backend test --chain-id feeappd-t1 -y   --memo "$MEMO"

#sleep 20  # wait for the roundtrip

#new_balances=$(feeappd query bank balances "$myaccount" -o json | jq '.balances')
#echo "old balances: $balances, new balances: $new_balances"

