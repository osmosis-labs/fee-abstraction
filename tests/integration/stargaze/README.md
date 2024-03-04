# Integrate `x/feeabs` with Stargaze

This is document for testing integrate a new chain with `x/feeabs` module.

Firstly, we will add `feeabs` module to Stargaze and using Osmosis testnet to minimize effort.

- Original Fee Abstraction repository:

  <https://github.com/osmosis-labs/fee-abstraction>

- Cosmos Hub with added `feeabs` module:

  <https://github.com/notional-labs/stargaze/tree/feature/feeabs>

## 1. Setup nodes

- Install Stargaze CLI(starsd):

```bash
git clone https://github.com/notional-labs/stargaze.git
cd gaia
git checkout feature/feeabs
make install
```

- Install Osmosis CLI:

```bash
git clone https://github.com/osmosis-labs/osmosis.git
cd osmosis
make install
```

- Run nodes:

  - `starsd`: In `stargaze` working dir, run:

  ```bash
  ./startnode.sh
  ```

- Enable `feeabs` by change params of this module
- Add host zone config proposal for `feeabs`

## 2. Setup relayers

```bash
#!/bin/sh

build/rly config init
build/rly chains add-dir <PATH_TO_CHAIN_CONFIG> // Example:  relayer/chains
build/rly paths add-dir <PATH_TO_PATHS_CONFIG_DIR> // Example: relayer/paths

build/rly keys restore stargaze relayer <STARGAZE_MNEMONIC\>
build/rly keys restore osmosis relayer  <OSMOSIS_MNEMONIC\>"


build/rly tx link transfer -d -t 10s
build/rly tx link feeabs --src_port feeabs --dst_port icqhost --order unordered --version icq-1build/rly start transfer
```

We should fund the wallets and reuse the wallets for relaying.

## 3. Setup pools, contracts on Osmosis testnet

### Pre-Deployed Contract on Osmosis testnet

owner: osmo1sg5ta3eaed5wxxpnq3u5463lysfg7ytjxqvk43

- Registry:

  - Code_id: 7238
  - Contract_addr: osmo1m9jk8zvrkpex0rxhp76emr0qm2z5khvj09msl9c78gcq7c38xdzsgq0cgm

- Swap:

  - Code_id: 7239
  - Contract_addr: osmo1j48ncj9wkzs3pnkux96ct6peg7rznnt4jx6ysdcs0283ysxj2ztqtr602y

- XCS:
  - code_id: 7240
  - Contract_addr: osmo177jurcy582fk5q298es6662pu48a46ze6eequnv3z0parekpwhhs034wsv
    We use the existing stored bytecode to instantiate the contract.

```bash
NODE=https://osmosis-testnet-rpc.polkachu.com:443
REG=./bytecode/crosschain_registry.wasm
# osmosisd tx wasm store $REG --from relayer --node $NODE --chain-id=osmo-test-5 --gas-prices 0.1uosmo --gas auto --gas-adjustment 1.3

REG_ID=7238
RELAYER_ADDR=osmo1sg5ta3eaed5wxxpnq3u5463lysfg7ytjxqvk43

# init registry contract
REG_INIT='{"owner": "'$RELAYER_ADDR'"}'
osmosisd tx wasm instantiate $REG_ID $REG_INIT --label "Registry Feeabs" --admin $RELAYER_ADDR --from relayer --node $NODE --chain-id=osmo-test-5 --gas-prices 0.1uosmo --gas auto --gas-adjustment 1.3

# init cross chain swap contract
XCS_ID=7240
XCS_INIT='{"swap_contract":"osmo1j48ncj9wkzs3pnkux96ct6peg7rznnt4jx6ysdcs0283ysxj2ztqtr602y","governor":"osmo1sg5ta3eaed5wxxpnq3u5463lysfg7ytjxqvk43", "registry_contract":"osmo1m9jk8zvrkpex0rxhp76emr0qm2z5khvj09msl9c78gcq7c38xdzsgq0cgm"}'
osmosisd tx wasm instantiate $XCS_ID $XCS_INIT --label "Registry Feeabs" --admin $RELAYER_ADDR --from relayer --node $NODE --chain-id=osmo-test-5 --gas-prices 0.1uosmo --gas auto --gas-adjustment 1.3

# Setup the pfm on stargaze for path unwinding
REG_ADDR=osmo1m9jk8zvrkpex0rxhp76emr0qm2z5khvj09msl9c78gcq7c38xdzsgq0cgm
PFM_EXEC='{"propose_pfm":{"chain": "stargaze"}}'
osmosisd tx wasm execute $REG_ADDR $PFM_EXEC --from relayer --amount 100000ibc/BD47A6048AA3BDC7E9DC98E0CB31AB777E6E8561E9D0BA45E13CA6EEB1558CB5 --node $NODE --chain-id=osmo-test-5 --gas-prices 0.1uosmo --gas auto --gas-adjustment 1.3

# check if properly set
QUERY='{"has_packet_forwarding": {"chain": "stargaze"}}'
osmosisd query wasm contract-state smart osmo1m9jk8zvrkpex0rxhp76emr0qm2z5khvj09msl9c78gcq7c38xdzsgq0cgm $QUERY_STAR --node $NODE

## data: true


```

After setting up the relayer, we should know the ibc denom of ustars and uosmo on each counterparty chain
For instance in this example:
starsOnOsmosis denom: ibc/BD47A6048AA3BDC7E9DC98E0CB31AB777E6E8561E9D0BA45E13CA6EEB1558CB5(transfer/channel-6061/ustars)
osmoOnStargaze denom: ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518(transfer/channel-0/osmo)

We create a pool on Osmosis testnet to swap between native denom(`ustars`) and external denom(in our case, `uosmo`)

```json
{
  "weights": "7uosmo,3ibc/BD47A6048AA3BDC7E9DC98E0CB31AB777E6E8561E9D0BA45E13CA6EEB1558CB5",
  "initial-deposit": "1000uosmo,200ibc/BD47A6048AA3BDC7E9DC98E0CB31AB777E6E8561E9D0BA45E13CA6EEB1558CB5",
  "swap-fee": "0.01",
  "exit-fee": "0",
  "future-governor": ""
}
```

```bash
osmosisd tx gamm create-pool --pool-file sample_pool.json --node $NODE --chain-id=osmo-test-5 --gas-prices 0.1uosmo --gas auto --gas-adjustment 1.3 --from relayer
```

Pool is now created, we can check the pool id based on the tx hash on explorer. In our case the pool_id value is 404

#### Setting up the swap router

We need to set route for swap.

```bash
SET_ROUTE='{"set_route":{"input_denom":"ibc/BD47A6048AA3BDC7E9DC98E0CB31AB777E6E8561E9D0BA45E13CA6EEB1558CB5", "output_denom":"uosmo", "pool_route":[{"pool_id": "404", "token_out_denom":"uosmo"}]}}'
osmosisd tx wasm execute osmo1j48ncj9wkzs3pnkux96ct6peg7rznnt4jx6ysdcs0283ysxj2ztqtr602y $SET_ROUTE --node $NODE --chain-id=osmo-test-5 --gas-prices 0.1uosmo --gas auto --gas-adjustment 1.3 --from relayer

SET_ROUTE='{"set_route":{"input_denom":"uosmo", "output_denom":"ibc/BD47A6048AA3BDC7E9DC98E0CB31AB777E6E8561E9D0BA45E13CA6EEB1558CB5", "pool_route":[{"pool_id": "404", "token_out_denom":"ibc/BD47A6048AA3BDC7E9DC98E0CB31AB777E6E8561E9D0BA45E13CA6EEB1558CB5"}]}}'
osmosisd tx wasm execute osmo1j48ncj9wkzs3pnkux96ct6peg7rznnt4jx6ysdcs0283ysxj2ztqtr602y $SET_ROUTE --node $NODE --chain-id=osmo-test-5 --gas-prices 0.1uosmo --gas auto --gas-adjustment 1.3 --from relayer
```

#### Setup host zone config

Then we create a proposla to setup hostzone config
proposal.json

```json
{
    {
    "title": "Add Fee Abbtraction Host Zone Proposal",
    "description": "Add Fee Abbtraction Host Zone",
    "host_chain_fee_abs_config": {
      "ibc_denom": "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518",
      "osmosis_pool_token_denom_in": "uosmo",
      "pool_id": "404",
      "frozen": false
    },
    "deposit": "100000000ustars"
  }
}
```

```bash
starsd tx gov submit-legacy-proposal add-hostzone-config proposal.json --from investor
starsd tx gov vote 2 yes  --from investor
starsd tx gov vote 2 yes  --from validator
```

#### Setup param changes

This is the final step for setting up the x/feeabs module. We need to configure some params, which we will modify via governance

```json
{
  "title": "Enable Fee Abtraction",
  "description": "Change params for enable fee abstraction",
  "changes": [
    {
      "subspace": "feeabs",
      "key": "NativeIbcedInOsmosis",
      "value": "ibc/BD47A6048AA3BDC7E9DC98E0CB31AB777E6E8561E9D0BA45E13CA6EEB1558CB5"
    },
    {
      "subspace": "feeabs",
      "key": "ChainName",
      "value": "stargaze"
    },
    {
      "subspace": "feeabs",
      "key": "IbcTransferChannel",
      "value": "channel-0"
    },
    {
      "subspace": "feeabs",
      "key": "IbcQueryIcqChannel",
      "value": "channel-1"
    },
    {
      "subspace": "feeabs",
      "key": "OsmosisCrosschainSwapAddress",
      "value": "osmo177jurcy582fk5q298es6662pu48a46ze6eequnv3z0parekpwhhs034wsv"
    }
  ],
  "deposit": "100000000ustars"
}
```

```bash
starsd tx gov submit-legacy-proposal param-change params.json --from investor --gas auto --gas-adjustment 1.3
starsd tx gov vote <PROPOSAL_ID> yes  --from investor
starsd tx gov vote <PROPOSAL_ID> yes  --from validator
```

## 4. Testing fee abstraction
