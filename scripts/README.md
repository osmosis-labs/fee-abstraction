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

feeappd tx gov vote 3 yes --from feeapp1 --keyring-backend test --chain-id feeappd-t1 --yes
```

```
feeappd tx gov submit-proposal param-change scripts/proposal_query.json --from feeacc --keyring-backend test --chain-id feeappd-t1 --yes

feeappd tx gov vote 1 yes --from feeapp1 --keyring-backend test --chain-id feeappd-t1 --yes

feeappd tx gov submit-proposal add-hostzone-config scripts/host_zone_query.json --from feeacc --keyring-backend test --chain-id feeappd-t1 --yes               

feeappd tx gov vote 2 yes --from feeapp1 --keyring-backend test --chain-id feeappd-t1 --yes
```

```
{ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518 ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878  channel-0  osmo1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqvlx82r 1 true true channel-1}
B834FA96EB41DB72C1DFA61DAE0000C76065ADAC
0ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518
```