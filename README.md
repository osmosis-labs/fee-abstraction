# Happy new year !!!!!!!
# Fee Abstraction
 
## Description 

Fee-abs enable users on any cosmos-sdk chain with ibc connections to pay fee using ibc token.

Fee-abs implememtation composes of 2 pieces of software :
 - Fee-abs module imported to the customer chain.
 - A smart contract deployed Osmosis. 

Other than those two, the implememtation also uses osmosis contract swap router which is already deployed on osmosis testnet. 

## Prototype

Firstly, we narrow the feature of fee-abs from allowing general ibc token as tx fee to allowing only ibc-ed osmosis as tx fee. If thing goes smoothly , we'll work on developing the full feature of fee-abs.

Fee-abs mechanism in a nutshell:
 1. Pulling `twap data` and update exchange rate: 
 - Periodically pulling `twap data` from osmosis using `ibc query`, this `twap data` will update the exchange rate of ibc-tokens to customer chain's native token. 
 2. Handling txs with ibc-token fee: 
 - The exchange rate is used to calculate the ammount of ibc-token needed for tx fee allowing users to pay ibc-token for tx fee instead of chain's native token.
 3. Swap accumulated ibc-tokens fee:
 - The collected ibc-tokens users use for tx fee is periodically swaped back to customer chain's native token using osmosis.

We'll goes into all the details now:

#### Pulling `twap data` and update exchange rate
    

#### Handling txs with ibc-token fee


#### Swap accumulated ibc-tokens fee



## Technical Architecture
![Overview](https://i.imgur.com/zFDI7Ce.png)
This is overview of fee abstraction workflow.


![Sequence diagram](https://i.imgur.com/cxTgDDh.png)


## Outcome

### Milestone 1. (Completed)
Deliver Osmosis TWAP oracle infrastructure.

We using a contract from Osmosis side to query TWAP over ibc. Fee Abstraction module send a IBC packet requires TWAP data to osmosis, and the contract will response TWAP data through ACK packet.

### Milestone 2  (Completed) 
Deliver modified gas fee SDK module that allows non-native token to be used for transaction fees. 

This module using `IBC Hooks` features in osmosis chain to swapping of accumulated fees. https://github.com/osmosis-labs/osmosis/tree/main/x/ibc-hooks


### Milestone 3: (In processing)

Deliver modified gas fee SDK module that allows non-native tokens to be used for transaction fees. This module would perform the swapping of accumulated fees using . 

Currently, we can only use ibc-ed osmosis token for transaction fees in Fee Abtraction chain. We're working for allow more IBC tokens to be used as transaction fees.


## Resources
 - Main repo: https://github.com/notional-labs/fee-abstraction
 - Contract repo: https://github.com/notional-labs/feeabstraction-contract