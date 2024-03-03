# Integrate `x/feeabs` with new chain

This is document for testing integrate a new chain with `x/feeabs` module.

Firstly, we will add `feeabs` module to Cosmos Hub, Stargaze and using Osmosis testnet to minimize effort.

* Original Fee Abstraction repository:

    <https://github.com/osmosis-labs/fee-abstraction>

* Cosmos Hub with added `feeabs` module:

    <https://github.com/notional-labs/gaia/tree/feeabs>

* Stargaze with added `feeabs` module:

    <https://github.com/notional-labs/stargaze/tree/feature/feeabs>

## 1. Setup nodes

* Install Gaia CLI
* Install Starsgaze CLI
* Install Osmosis CLI
* Run nodes:
  * Gaiad
  * Starsd
* Enable `feeabs` by change params of this module
* Add host zone config proposal for `feeabs`

## 2. Setup relayers

## 3. Setup pools, contracts on Osmosis testnet

## 4. Testing fee abstraction
