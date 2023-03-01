## Abstract

When making a transactions, usually users need to pay fees in native token, but `Feeabs` module enable users on any Cosmos chain which implements this module with IBC connections to pay fee using ibc token. When users use an ibc denom as fees, the ``FeeAbstrationMempoolFeeDecorator`` ante handler will check whether the chain support the transactions to be paid by that ibc denom. It will calculate the amount of ibc tokens equivalent to native token when users make a normal transaction based on Osmosis ``twap`` between ibc denom and native denom.

After that, the ``FeeAbstractionDeductFeeDecorate`` ante handler swaps ibc token for native token to pay for transaction fees. The accumulated ibc token will be swaped on Osmosis Dex every epoch.

The `Feeabs` module fetches Osmosis [twap](https://github.com/osmosis-labs/osmosis/tree/main/x/twap) at the begin of every [epoch](01_concepts.md#Epoch) and swap all of ibc tokens left in the module to native token using `swap router` and `ibc hooks` on Osmosis.

## Contents

1. **[Concepts](01_concepts.md)**
   - [epoch](01_concepts.md#Epoch)
2. **[State](02_state.md)**
   - [OsmosisTwapExchangeRate](02_state.md#OsmosisTwapExchangeRate)
   - [HostChainChainConfig](02_state.md#HostChainChainConfig)
3. **[Epoch](03_epoch.md)**
4. **[Events](04_events.md)**
   - [BeginBlocker](04_events.md#BeginBlocker)
   - [IBC](04_events.md#IBC)
5. **[Parameters](05_params.md)**

## Interagation to 