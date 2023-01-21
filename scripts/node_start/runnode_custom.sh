#!/bin/bash
# Run this script to quickly install, setup, and run the current version of juno without docker.
# ./scripts/test_node.sh [clean|c]

KEY="feeapp1"
CHAINID="feeappd-t1"
MONIKER="localfeeappd"
KEYALGO="secp256k1"
KEYRING="test"
LOGLfeeappdL="info"

feeappd config keyring-backend $KEYRING
feeappd config chain-id $CHAINID

command -v feeappd > /dev/null 2>&1 || { echo >&2 "feeappd command not found. Ensure this is setup / properly installed in your GOPATH."; exit 1; }
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

from_scratch () {

  make install

  # remove existing daemon.
  rm -rf ~/.feeappd/*

  # juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk
  echo "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry" | feeappd keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --recover
  # juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl
  echo "wealth flavor believe regret funny network recall kiss grape useless pepper cram hint member few certain unveil rather brick bargain curious require crowd raise" | feeappd keys add feeacc --keyring-backend $KEYRING --algo $KEYALGO --recover

  feeappd init $MONIKER --chain-id $CHAINID

  # Function updates the config based on a jq argument as a string
  update_test_genesis () {
    # update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
    cat $HOME/.feeappd/config/genesis.json | jq "$1" > $HOME/.feeappd/config/tmp_genesis.json && mv $HOME/.feeappd/config/tmp_genesis.json $HOME/.feeappd/config/genesis.json
  }

  # Set gas limit in genesis
  update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
  update_test_genesis '.app_state["gov"]["voting_params"]["voting_period"]="15s"'

  update_test_genesis '.app_state["staking"]["params"]["bond_denom"]="stake"'
  update_test_genesis '.app_state["bank"]["params"]["send_enabled"]=[{"denom": "stake","enabled": true}]'
  # update_test_genesis '.app_state["staking"]["params"]["min_commission_rate"]="0.100000000000000000"' # sdk 46 only

  update_test_genesis '.app_state["mint"]["params"]["mint_denom"]="stake"'
  update_test_genesis '.app_state["gov"]["deposit_params"]["min_deposit"]=[{"denom": "stake","amount": "1000000"}]'
  update_test_genesis '.app_state["crisis"]["constant_fee"]={"denom": "stake","amount": "1000"}'

  update_test_genesis '.app_state["tokenfactory"]["params"]["denom_creation_fee"]=[{"denom":"stake","amount":"100"}]'

  update_test_genesis '.app_state["feeshare"]["params"]["allowed_denoms"]=["stake"]'

  # Allocate genesis accounts
  feeappd add-genesis-account $KEY 10000000000000stake,100000000000000utest --keyring-backend $KEYRING
  feeappd add-genesis-account feeacc 10000000000000stake,100000000000000utest --keyring-backend $KEYRING

  feeappd gentx $KEY 10000000000000stake --keyring-backend $KEYRING --chain-id $CHAINID

  # Collect genesis tx
  feeappd collect-gentxs

  # Run this to ensure junorything worked and that the genesis file is setup correctly
  feeappd validate-genesis
}


if [ $# -eq 1 ] && [ $1 == "clean" ] || [ $1 == "c" ]; then
  echo "Starting from a clean state"
  from_scratch
fi

echo "Starting node..."

# Opens the RPC endpoint to outside connections
sed -i '/laddr = "tcp:\/\/127.0.0.1:26657"/c\laddr = "tcp:\/\/0.0.0.0:26657"' ~/.feeappd/config/config.toml
sed -i 's/cors_allowed_origins = \[\]/cors_allowed_origins = \["\*"\]/g' ~/.feeappd/config/config.toml
sed -i 's/enable = false/enable = true/g' ~/.feeappd/config/app.toml
sed -i '/address = "tcp:\/\/0.0.0.0:1317"/c\address = "tcp:\/\/0.0.0.0:1318"' ~/.feeappd/config/app.toml

feeappd config node tcp://127.0.0.1:2241
feeappd start --pruning=nothing  --minimum-gas-prices=0stake --p2p.laddr tcp://0.0.0.0:2240 --rpc.laddr tcp://127.0.0.1:2241 --grpc.address 127.0.0.1:2242 --grpc-web.address 127.0.0.1:2243