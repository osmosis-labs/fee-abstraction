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
  * `Gaiad`: In `gaia` working dir, run:

  ```bash
  sh startnode.sh
  ```

* Enable `feeabs` by change params of this module
* Add host zone config proposal for `feeabs`

## 2. Setup relayers

## 3. Setup pools, contracts on Osmosis testnet

## 4. Testing fee abstraction
