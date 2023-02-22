# Fee Abstraction
 
## Description 

Fee-abs enable users on any cosmos-sdk chain with ibc connections to pay fee using ibc token.

Fee-abs implememtation composes of 2 pieces of software :
 - Fee-abs module imported to the customer chain.
 - A smart contract for querying twap deployed Osmosis. 

Other than those two, the implememtation also uses osmosis contract swap router which is already deployed on osmosis testnet. 

## Prototype

Firstly, we narrow the feature of fee-abs from allowing general ibc token as tx fee to allowing only ibc-ed osmosis as tx fee. If thing goes smoothly , we'll work on developing the full feature of fee-abs.

Fee-abs mechanism in a nutshell:
 1. Pulling `twap data` and update exchange rate: 
 - Periodically pulling `twap data` from osmosis by ibc-ing to our osmosis contract, this `twap data` will update the exchange rate of osmosis to customer chain's native token. 
 2. Handling txs with ibc-osmosis fee: 
 - The exchange rate is used to calculate the ammount of ibc-osmosis needed for tx fee allowing users to pay ibc-osmosis for tx fee instead of chain's native token.
 3. Swap accumulated ibc-osmosis fee:
 - The collected ibc-osmosis users use for tx fee is periodically swaped back to customer chain's native token using osmosis.

We'll goes into all the details now:

#### Pulling `twap data` and update exchange rate
For this to work, we first has to set up an ibc channel from fee-abs to our osmosis contract. This channel set-up process can be done by anyone, just like setting up an ibc transfer channel. Once that ibc channel is there, we'll use that channel to ibc-query Twap data. Let's call this the querying channel.

The process of pulling Twap data and update exchange rate :

![](https://i.imgur.com/sJA4yV7.png)

Description :
    For every `update exchange rate period`, at fee-abs `EndBlocker()` we submit a `query twap packet` to the querying channel on the customer chain's end. Then relayers will submit `MsgReceivePacket` so that our `QueryTwapPacket` which will be routed to our osmosis contract to be processed. Our osmosis contract then query twap price and put it in the ibc acknowledgement. Relayers then submit `MsgAcknowledgement` to the customer chain so that the ibc acknowledgement is routed to fee-abs to be processed. Fee-abs then update exchange rate according to the Twap wrapped in the ibc acknowledgement.

#### Handling txs with ibc-token fee
We modified `MempoolFeeDecorator` so that it can handle ibc-osmosis as fee. If the tx has osmosis fee, we basically replace the ibc-osmosis ammount with the equivalent native-token ammount which is calculated by `exchange rate` * `ibc-osmosis ammount`.

We have an account to manage the ibc-osmosis user used to pay for tx fee. The collected osmosis fee is sent to that account instead of community pool account.

#### Swap accumulated ibc-tokens fee
We use osmosis's ibc hook feature to do this. We basically ibc transfer to the osmosis crosschain swap contract with custom memo to swap the osmosis fee back to customer chain's native-token and ibc transfer back to the customer chain.

## Resources
 - Main repo: https://github.com/notional-labs/fee-abstraction
 - Contract repo: https://github.com/notional-labs/feeabstraction-contract

```
{"wasm":{"contract":"osmo1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqvlx82r","msg":{"osmosis_swap":{"input_coin":{"denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","amount":"1000"},"output_denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","slippage":{"twap":{"slippage_percentage":"20","window_seconds":10}},"receiver":"feeabs1hj5fveer5cjtn4wd6wstzugjfdxzl0xpjhy828"}},"receiver":"feeabs1hj5fveer5cjtn4wd6wstzugjfdxzl0xpjhy828"}}
```