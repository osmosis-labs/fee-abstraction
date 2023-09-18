# Test feeabs and osmosis Interchain Query for spot price

## Setup

# Deploy chains
Run Feeabs node
```
./scripts/node_start/runnode_custom.sh c
```
Run Osmosis node
```
./scripts/node_start/runnode_osmosis.sh c
```
Run Gaia node
```
./scripts/node_start/runnode_gaia.sh c
```
# Run relayer
Run Feeabs - Osmosis relayer
```
./scripts/run_relayer_feeabs_osmo.sh
```
Run Feeabs - Gaia relayer
```
./scripts/run_relayer_feeabs_gaia.sh
```
Run Osmosis - Gaia relayer
```
./scripts/run_relayer_osmo_gaia.sh
```
# Setup IBC channel
Run the following command to setup IBC channel
```
hermes --config scripts/relayer_hermes/config_feeabs_osmosis.toml create channel --a-chain testing --b-chain feeappd-t1 --a-port transfer --b-port transfer --new-client-connection --yes
hermes --config scripts/relayer_hermes/config_feeabs_osmosis.toml create channel --a-chain testing --b-chain feeappd-t1 --a-port icqhost --b-port feeabs --new-client-connection --yes
hermes --config scripts/relayer_hermes/config_feeabs_gaia.toml create channel --a-chain gaiad-t1 --b-chain feeappd-t1 --a-port transfer --b-port transfer --new-client-connection --yes
hermes --config scripts/relayer_hermes/config_osmosis_gaia.toml create channel --a-chain gaiad-t1 --b-chain testing --a-port transfer --b-port transfer --new-client-connection --yes
```

After running, the channel map should like this
```
feeabs - osmo: channel-0 channel-0
feeabs - osmo: (feeabs - icqhost) channel-1 channel-1
feeabs - gaia: channel-2 channel-0
osmo   - gaia: channel-2 channel-1
```

# Create an osmosis pool
Get Osmosis testing address
```
export VALIDATOR=$(osmosisd keys show validator1 -a --keyring-backend test)
export OWNER=$(osmosisd keys show deployer -a --keyring-backend test)
```

Transfer token from Feeabs and Gaia to Osmosis
```
feeappd tx ibc-transfer transfer transfer channel-0 "$VALIDATOR" 1000000000000stake --from feeacc --keyring-backend test --chain-id feeappd-t1 --yes --fees 5000stake
gaiad tx ibc-transfer transfer transfer channel-1 "$VALIDATOR" 1000000000000uatom --from gnad --keyring-backend test --chain-id gaiad-t1 --yes --fees 5000stake
```

Create pool
```
cat > sample_pool.json <<EOF
{
        "weights": "1ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E,1ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878",
        "initial-deposit": "500000000000ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E,100000000000ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878",
        "swap-fee": "0.01",
        "exit-fee": "0",
        "future-governor": "168h"
}
EOF

osmosisd tx gamm create-pool --pool-file sample_pool.json --from validator1 --keyring-backend=test --home=$HOME/.osmosisd --chain-id testing --yes --fees 5000stake --gas 400000
```

# Deploy contract and create relayer channel
```./scripts/ibc_swap/deploy_osmosis_contract.sh```


## Gov proposal

```
feeappd tx gov submit-proposal param-change scripts/proposal.json --from feeacc --keyring-backend test --chain-id feeappd-t1 --yes

feeappd tx gov vote 1 yes --from feeapp1 --keyring-backend test --chain-id feeappd-t1 --yes

feeappd tx gov submit-proposal add-hostzone-config scripts/host_zone_gaia.json --from feeacc --keyring-backend test --chain-id feeappd-t1 --yes               

feeappd tx gov vote 2 yes --from feeapp1 --keyring-backend test --chain-id feeappd-t1 --yes
```

## Fund module account
```
feeappd tx feeabs fund 500000000stake --from myaccount --keyring-backend test --chain-id feeappd-t1 -y
```

## Test
```
feeappd tx feeabs query-osmosis-twap --from myaccount --keyring-backend test --chain-id feeappd-t1 --yes --fees 5000stake
# Wait for about 10 sec
feeappd q feeabs osmo-arithmetic-twap ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E
```

The result response twap of ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E (uatom)
Now you can pay fees using ```ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E```
