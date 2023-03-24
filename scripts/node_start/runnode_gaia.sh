#!/bin/bash
# Run this script to quickly install, setup, and run the current version of juno without docker.
# ./scripts/test_node.sh [clean|c]

KEY="gaiaapp1"
CHAINID="gaiad-t1"
MONIKER="localgaiad"
KEYALGO="secp256k1"
KEYRING="test"

gaiad config keyring-backend $KEYRING
gaiad config chain-id $CHAINID

command -v gaiad > /dev/null 2>&1 || { echo >&2 "gaiad command not found. Ensure this is setup / properly installed in your GOPATH."; exit 1; }
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

from_scratch () {

  make install

  # remove existing daemon.
  rm -rf ~/.gaia/*

  # juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk
  echo "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry" | gaiad keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --recover
  # juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl
  echo "cup pencil conduct depth analyst human trick excite gain copy option arena mix stamp team soon embody jewel erupt advice access prefer negative cost" | gaiad keys add gnad --keyring-backend $KEYRING --algo $KEYALGO --recover

  gaiad init $MONIKER --chain-id $CHAINID

  # Function updates the config based on a jq argument as a string
  update_test_genesis () {
    # update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
    cat $HOME/.gaia/config/genesis.json | jq "$1" > $HOME/.gaia/config/tmp_genesis.json && mv $HOME/.gaia/config/tmp_genesis.json $HOME/.gaia/config/genesis.json
  }

  # Set gas limit in genesis
  update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
  update_test_genesis '.app_state["gov"]["voting_params"]["voting_period"]="45s"'

  update_test_genesis '.app_state["staking"]["params"]["bond_denom"]="stake"'
  update_test_genesis '.app_state["bank"]["params"]["send_enabled"]=[{"denom": "stake","enabled": true}]'
  # update_test_genesis '.app_state["staking"]["params"]["min_commission_rate"]="0.100000000000000000"' # sdk 46 only

  update_test_genesis '.app_state["mint"]["params"]["mint_denom"]="stake"'
  update_test_genesis '.app_state["gov"]["deposit_params"]["min_deposit"]=[{"denom": "stake","amount": "1000000"}]'
  update_test_genesis '.app_state["crisis"]["constant_fee"]={"denom": "stake","amount": "1000"}'

  update_test_genesis '.app_state["tokenfactory"]["params"]["denom_creation_fee"]=[{"denom":"stake","amount":"100"}]'

  update_test_genesis '.app_state["feeshare"]["params"]["allowed_denoms"]=["stake"]'

  # Allocate genesis accounts
  gaiad add-genesis-account $KEY 10000000000000uatom,10000000000000stake,100000000000000utest --keyring-backend $KEYRING
  gaiad add-genesis-account gnad 10000000000000uatom,10000000000000stake,100000000000000utest --keyring-backend $KEYRING
  
  gaiad gentx $KEY 10000000000000stake --keyring-backend $KEYRING --chain-id $CHAINID

  # Collect genesis tx
  gaiad collect-gentxs

  # Run this to ensure junorything worked and that the genesis file is setup correctly
  gaiad validate-genesis
}


if [ $# -eq 1 ] && [ $1 == "clean" ] || [ $1 == "c" ]; then
  echo "Starting from a clean state"
  from_scratch
fi

echo "Starting node..."

gaiad config node tcp://0.0.0.0:3241
gaiad start --pruning=nothing  --minimum-gas-prices=0stake --p2p.laddr tcp://0.0.0.0:3240 --rpc.laddr tcp://0.0.0.0:3241 --grpc.address 0.0.0.0:3242 --grpc-web.address 0.0.0.0:3243