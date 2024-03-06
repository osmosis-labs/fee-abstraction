# Integrate `x/feeabs` with Cosmos Hub

This is document for testing integrate a new chain with `x/feeabs` module.

Firstly, we will add `feeabs` module to Cosmos Hub and using Osmosis testnet to minimize effort.

* Original Fee Abstraction repository:

    <https://github.com/osmosis-labs/fee-abstraction>

* Cosmos Hub with added `feeabs` module:

    <https://github.com/notional-labs/gaia/tree/feeabs>

## 1. Setup nodes

* Install Gaia CLI:

```bash
git clone https://github.com/notional-labs/gaia.git
cd gaia
git checkout feeabs
make install
```

* Install Osmosis CLI:

```bash
git clone https://github.com/osmosis-labs/osmosis.git
cd osmosis
make install
```

* Run nodes:
* In `gaia` working dir, run:

```bash
sh startnode.sh
```

## 2. Setup relayers

* Run relayers:

```bash
# Setup chains and paths for relayer between gaia-feeabs
cd <fee-abstraction>/tests/integration/cosmoshub
sh tools/setup_relayer.sh

rly tx link transfer -d -t 10s --client-tp 36h
rly tx link query -d -t 10s --client-tp 36h

rly start
```

* Get ibc token info:

```bash
# Get channels in gaia
rly q channels gaia

# {"chain_id":"gaia-1","channel_id":"channel-0","client_id":"07-tendermint-0","connection_hops":["connection-0"],"counterparty":{"chain_id":"osmo-test-5","channel_id":"channel-6084","client_id":"07-tendermint-2545","connection_id":"connection-2390","port_id":"transfer"},"ordering":"ORDER_UNORDERED","port_id":"transfer","state":"STATE_OPEN","version":"ics20-1"}
# {"chain_id":"gaia-1","channel_id":"channel-1","client_id":"07-tendermint-0","connection_hops":["connection-1"],"counterparty":{"chain_id":"osmo-test-5","channel_id":"channel-6085","client_id":"07-tendermint-2545","connection_id":"connection-2391","port_id":"transfer"},"ordering":"ORDER_UNORDERED","port_id":"transfer","state":"STATE_OPEN","version":"ics20-1"}

# Transfer uatom from gaia to osmosis
rly tx transfer gaia osmosis 100000000uatom osmo1wrhdsm4gy307mygkgmanjc3r2g0ttuhnhkfp44 channel-0 --path transfer

# Transfer uosmo from osmosis to gaia
rly tx transfer osmosis gaia 100000uosmo cosmos1wrhdsm4gy307mygkgmanjc3r2g0ttuhnld63r8 channel-6084 --path transfer


# Query balances of relayer wallet
rly q balance gaia --ibc-denoms
rly q balance osmosis --ibc-denoms

# ibc/uatom on osmosis: ibc/80C64E7EB7E8B6705FC9C1D9C486EB6278823068D9224915B6A5DABDF03FB2D5
# ibc/uosmo on gaia: ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518
```

## 3. Setup pools, contracts on Osmosis testnet

a. Pre-deploy contracts:

* Registry:

  * code_id: `7238`
  * address: `osmo1m9jk8zvrkpex0rxhp76emr0qm2z5khvj09msl9c78gcq7c38xdzsgq0cgm`

* SwapRouter:

  * code_id: `7239`
  * address: `osmo1j48ncj9wkzs3pnkux96ct6peg7rznnt4jx6ysdcs0283ysxj2ztqtr602y`

* XCSv2:
  * code_id: `7240`
  * address: `osmo177jurcy582fk5q298es6662pu48a46ze6eequnv3z0parekpwhhs034wsv`

b. Create pool:

Correct information on `<cosmoshub>/pools/pool.json`, and run script:

```bash
sh tools/create_pool.sh
```

Query `txhash` and find `pool_id`:

```bash
pool_id=409 
```

c. Set swap route

Correct information on `<cosmoshub>/tools/set_route.sh`, and run script:

```bash
sh tools/set_route.sh
```

## 4. Testing fee abstraction

a. Change params of `feeabs` module

Correct information on `<cosmoshub>/proposals/params.json`, and run script:

```bash
sh tools/change_feeabs_params.sh
```

b. Add host zone config

Correct information on `<cosmoshub>/proposals/add_host_zone.json`, and run script:

```bash
sh tools/add_host_zone.sh
```
