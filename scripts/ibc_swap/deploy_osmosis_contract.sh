#!/usr/bin/env bash

CHANNEL_ID="channel-0"
export VALIDATOR=$(osmosisd keys show validator1 -a --keyring-backend test )
echo $VALIDATOR=$(osmosisd keys show validator1 -a --keyring-backend test )

export OWNER=$(osmosisd keys show deployer -a --keyring-backend test )
echo $OWNER=$(osmosisd keys show deployer -a --keyring-backend test )


#store the SetupCrosschainRegistry
osmosisd tx wasm store scripts/bytecode/crosschain_registry.wasm --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --gas 10000000 --fees 25000stake --yes
sleep 5

# instantiate
INIT_SWAPREGISTRY='{"owner":"'$OWNER'"}'
osmosisd tx wasm instantiate 1 "$INIT_SWAPREGISTRY" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --label "test" --no-admin --yes --fees 5000stake
SWAPREGISTRY_ADDRESS=osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9
sleep 5

# execute
EXE_MSG='{"modify_chain_channel_links": {"operations": [{"operation": "set","source_chain": "feeappd-t1","destination_chain": "osmosis","channel_id": "channel-0"},{"operation": "set","source_chain": "osmosis","destination_chain": "feeappd-t1","channel_id": "channel-0"},{"operation": "set","source_chain": "feeappd-t1","destination_chain": "gaiad-t1","channel_id": "channel-2"},{"operation": "set","source_chain": "gaiad-t1","destination_chain": "feeappd-t1","channel_id": "channel-0"},{"operation": "set","source_chain": "osmosis","destination_chain": "gaiad-t1","channel_id": "channel-2"},{"operation": "set","source_chain": "gaiad-t1","destination_chain": "osmosis","channel_id": "channel-1"}]}}'
osmosisd tx wasm execute osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9 "$EXE_MSG" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --yes --fees 5000stake
sleep 5
PREFIX='{"modify_bech32_prefixes": {"operations": [{"operation": "set", "chain_name": "feeappd-t1", "prefix": "feeabs"},{"operation": "set", "chain_name": "osmosis", "prefix": "osmo"},{"operation": "set", "chain_name": "gaiad-t1", "prefix": "cosmos"}]}}'
osmosisd tx wasm execute osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9 "$PREFIX" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --yes --fees 5000stake
sleep 5
FEEABS_PFM='{"propose_pfm":{"chain": "feeappd-t1"}}'
osmosisd tx wasm execute osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9 "$FEEABS_PFM" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --yes --fees 5000stake
sleep 5
GAIA_PFM='{"propose_pfm":{"chain": "gaiad-t1"}}'
osmosisd tx wasm execute osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9 "$GAIA_PFM" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --yes --fees 5000stake
sleep 5

# Store the swaprouter contract
osmosisd tx wasm store scripts/bytecode/swaprouter.wasm --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --gas 10000000 --fees 25000stake --yes
sleep 5

# Instantiate the swaprouter contract
INIT_SWAPROUTER='{"owner":"'$OWNER'"}'
osmosisd tx wasm instantiate 2 "$INIT_SWAPROUTER" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --label "test" --no-admin --yes --fees 5000stake
sleep 5
SWAPROUTER_ADDRESS=osmo1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqvlx82r
echo $SWAPROUTER_ADDRESS

# Configure the swaprouter
CONFIG_SWAPROUTER='{"set_route":{"input_denom":"ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E","output_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","pool_route":[{"pool_id":"1","token_out_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"}]}}'
osmosisd tx wasm execute $SWAPROUTER_ADDRESS "$CONFIG_SWAPROUTER" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing -y --fees 5000stake
sleep 5

# Store the crosschainswap contract
osmosisd tx wasm store scripts/bytecode/crosschain_swaps.wasm --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --gas 10000000 --fees 25000stake --yes
sleep 10

# Instantiate the crosschainswap contract
INIT_CROSSCHAIN_SWAPS='{"swap_contract":"'$SWAPROUTER_ADDRESS'","governor": "'$OWNER'"}'
osmosisd tx wasm instantiate 3 "$INIT_CROSSCHAIN_SWAPS" --keyring-backend=test --home=$HOME/.osmosisd --from deployer --chain-id testing --label "test" --no-admin --yes --fees 5000stake
sleep 5
CROSSCHAIN_SWAPS_ADDRESS=osmo17p9rzwnnfxcjp32un9ug7yhhzgtkhvl9jfksztgw5uh69wac2pgs5yczr8
1