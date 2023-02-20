# Test feeabs and osmosis Interchain Query for spot price

## Setup
```
# Deploy chains
./scripts/node_start/runnode_custom.sh c
./scripts/node_start/runnode_osmosis.sh
# Run relayer
./scripts/run_relayer.sh
# Create an osmosis pool
./scripts/create_pool.sh
# Deploy contract and create relayer channel
./scripts/deploy_and_channel.sh
```

## Test
```
feeappd tx feeabs queryomosis --from feeacc --keyring-backend test --chain-id feeappd-t1 --yes
# Wait for about 10 sec
feeappd q feeabs osmo-spot-price
```

The result looks like this 
```
base_asset: osmo
quote_asset: stake
spot_price: "2.000000000000000000"
```

## Gov proposal

```
feeappd tx gov submit-proposal param-change scripts/proposal.json --from feeacc --keyring-backend test --chain-id feeappd-t1 --yes

feeappd tx gov vote 1 yes --from feeapp1 --keyring-backend test --chain-id feeappd-t1 --yes

feeappd tx gov submit-proposal add-hostzone-config scripts/host_zone.json --from feeacc --keyring-backend test --chain-id feeappd-t1 --yes               

feeappd tx gov vote 2 yes --from feeapp1 --keyring-backend test --chain-id feeappd-t1 --yes

```