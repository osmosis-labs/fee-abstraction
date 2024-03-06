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

* Run script:

```bash
cd <fee-abstraction>/tests/integration/cosmoshub
sh tools/run_relayer.sh
```

## 3. Setup pools, contracts on Osmosis testnet

Pre-deploy contracts:

* Registry:

  * code_id: `7238`
  * address: `osmo1m9jk8zvrkpex0rxhp76emr0qm2z5khvj09msl9c78gcq7c38xdzsgq0cgm`

* SwapRouter:

  * code_id: `7239`
  * address: `osmo1j48ncj9wkzs3pnkux96ct6peg7rznnt4jx6ysdcs0283ysxj2ztqtr602y`

* XCSv2:
  * code_id: `7240`
  * address: `osmo177jurcy582fk5q298es6662pu48a46ze6eequnv3z0parekpwhhs034wsv`

## 4. Testing fee abstraction
