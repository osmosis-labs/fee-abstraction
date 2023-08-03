KEY="mykey"
CHAINID="test-1"
MONIKER="localtestnet"
KEYALGO="secp256k1"
KEYRING="test"
LOGLEVEL="info"
# to trace evm
#TRACE="--trace"
TRACE=""

# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

# remove existing daemon
rm -rf ~/.feeapp*

feeappd config keyring-backend $KEYRING
feeappd config chain-id $CHAINID

# if $KEY exists it should be deleted
feeappd keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO

# Set moniker and chain-id for Evmos (Moniker can be anything, chain-id must be an integer)
feeappd init $MONIKER --chain-id $CHAINID 

# Allocate genesis accounts (cosmos formatted addresses)
feeappd add-genesis-account $KEY 100000000000000000000000000stake --keyring-backend $KEYRING

# Sign genesis transaction
feeappd gentx $KEY 1000000000000000000000stake --keyring-backend $KEYRING --chain-id $CHAINID

# Collect genesis tx
feeappd collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
feeappd validate-genesis

if [[ $1 == "pending" ]]; then
  echo "pending mode is on, please wait for the first block committed."
fi

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
feeappd start --pruning=nothing  --minimum-gas-prices=0.0001stake --rpc.laddr tcp://127.0.0.1:26650

