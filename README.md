# Fee Abstraction

## Context

The concrete use cases which motivated this module include:

- The desire to use IBC token as transaction fees on any chain instead of having to use native token as fee.
- To fully take advantage of the newly represented Osmosis [``swap router``](https://github.com/osmosis-labs/osmosis/tree/main/cosmwasm/contracts) with the [``ibc-hooks``](https://github.com/osmosis-labs/osmosis/tree/main/x/ibc-hooks) module and the [``async-icq``](https://github.com/strangelove-ventures/async-icq) module.


## Description

Fee abstraction modules enable users on any Cosmos chain with IBC connections to pay fee using ibc token.

Fee-abs implementation:

- Fee-abs module imported to the customer chain.

The implememtation also uses Osmosis swap router and async-icq module which are already deployed on Osmosis testnet.

## Prototype

Fee-abs mechanism in a nutshell:

 1. Pulling `twap data` and update exchange rate:

- Periodically pulling `twap data` from osmosis by ibc-ing to `async-icq` module on Osmosis, this `twap data` will update the exchange rate of osmosis to customer chain's native token.

 2. Handling txs with ibc-token fee:

- The exchange rate is used to calculate the amount of ibc-token needed for tx fee allowing users to pay ibc-token for tx fee instead of chain's native token.

 3. Swap accumulated ibc-token fee:

- The collected ibc-token users use for tx fee is periodically swaped back to customer chain's native token using osmosis.

We'll goes into all the details now:

#### Pulling `twap data` and update exchange rate

For this to work, we first has to set up an ibc channel from `feeabs` to `async-icq`. This channel set-up process can be done by anyone, just like setting up an ibc transfer channel. Once that ibc channel is there, we'll use that channel to ibc-query Twap data. Let's call this the querying channel.

The process of pulling Twap data and update exchange rate :

![Diagram of the process of pulling Twap data and updating exchange rate](https://i.imgur.com/HJ9a26H.png "Diagram of the process of pulling Twap data and updating exchange rate")

Description :
    For every `update exchange rate period`, at fee-abs `BeginBlocker()` we submit a `InterchainQueryPacketData` which wrapped `QueryArithmeticTwapToNowRequest` to the querying channel on the customer chain's end. Then relayers will submit `MsgReceivePacket` so that our `QueryTwapPacket` which will be routed to `async-icq` module to be processed. `async-icq` module then unpack `InterchainQueryPacketData` and send query to TWAP module. The correspone response will be wrapped in the ibc acknowledgement. Relayers then submit `MsgAcknowledgement` to the customer chain so that the ibc acknowledgement is routed to fee-abs to be processed. Fee-abs then update exchange rate according to the Twap wrapped in the ibc acknowledgement.

#### Handling txs with ibc-token fee

We modified `MempoolFeeDecorator` so that it can handle ibc-token as fee. If the tx has ibc-token fee, the AnteHandler will first check if that token is allowed (which is setup by Gov) we basically replace the amount of ibc-token with the equivalent native-token amount which is calculated by `exchange rate` * `ibc-token amount`.

We have an account to manage the ibc-token user used to pay for tx fee. The collected ibc-token fee is sent to that account instead of community pool account.

#### Swap accumulated ibc-tokens fee

Fee-abstraction will use osmosis's Cross chain Swap (XCS) feature to do this. We basically ibc transfer to the osmosis crosschain swap contract with custom memo to swap the osmosis fee back to customer chain's native-token and ibc transfer back to the customer chain.

##### How XCS work

###### Reverse With Path-unwinding to get Ibc-token on Osmosis

- Create a ibc transfer message with a specific MEMO to work with ibc [``packet-forward-middleware``](https://github.com/strangelove-ventures/packet-forward-middleware) which is path-unwinding (an ibc feature that allow to automatic define the path and ibc transfer multiple hop follow the defined path)
- Ibc transfer the created packet to get the fee Ibc-token on Osmosis

Ex: When you sent STARS on Hub to Osmosis, you will get Osmosis(Hub(STARS)) which is different with STARS on Osmosis Osmosis(STARS). It will reverse back Osmosis(Hub(STARS)) to Osmosis(STARS):

![Diagram of the process of swapping accumulated ibc-tokens fee](https://i.imgur.com/D1wSrMm.png "Diagram of the process of swapping accumulated ibc-tokens fee")

###### Swap Ibc-token

###### Swap Ibc-token

After reverse the ibc-token, XCS will :

- Swap with the specific pool (which is defined in the transfer packet from Feeabs-chain) to get Feeabs-chain native-token
- Transfer back Feeabs-chain native-token to Feeabs module account (will use to pay fee for other transaction)

![Diagram of the process of swapping accumulated ibc-tokens fee](https://i.imgur.com/YKOK8mr.png "Diagram of the process of swapping accumulated ibc-tokens fee")

Current version of fee-abstraction working with XCSv2


## Repository Structure

This repository is branched by the cosmos-sdk versions and ibc-go versions used.  Currently fee abstraction supports:

- SDK v0.50.x & IBC-go v8.*
  - note: incomplete
  - branch: release/v8.0.x
  - path: github.com/osmosis-labs/fee-abstraction/v8 
- SDK v0.47.x & IBC-go v7.*
  - branch: release/v7.0.x
  - path: github.com/osmosis-labs/fee-abstraction/v7
- SDK v0.45.x
  - branch: release/v4.0.x
  - path: github.com/osmosis-labs/fee-abstraction/v4
 
**note:** there is an sdk v0.46.x branch, but I don't recommend using it at this time. 

