## Context

The concrete use cases which motivated this module include:

- The desire to use IBC token as transaction fees on any chain instead of having to use native token as fee.
- To fully take advantage of the newly represented Osmosis [``swap router``](https://github.com/osmosis-labs/osmosis/tree/main/cosmwasm/contracts) with the [``ibc-hooks``](https://github.com/osmosis-labs/osmosis/tree/main/x/ibc-hooks) module.

## Description

Fee abstraction modules enable users on any Cosmos chain with IBC connections to pay fee using ibc token.

Fee-abs implementation composes of 2 pieces of software :

- Fee-abs module imported to the customer chain.
- A smart contract for querying TWAP deployed on Osmosis.

Other than those two, the implememtation also uses Osmosis swap router which is already deployed on Osmosis testnet.
