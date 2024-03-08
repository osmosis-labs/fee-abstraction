# Fee abstraction

1. Deduct fee
How it works:
* receive a transaction with non - native fees
* calculate the equivalent amount of native fees based on IBC query result of Osmosis TWAP prices
* send non - native fees from fee payer to feeabs module
* deduct native fees from feeabs module

SECURITY: 
* Assume that Osmosis TWAP prices are up - to - date to minimize slippage. A way to deal with slippage is that fee abstraction will only perform the swap of non - native tokens only when market prices are higher than DCA prices instead of time - based. Thus, FROZEN or OUTDATED state of host zone connection will be skipped.
* Assume that FeeAbs always have enough native fees to process transactions

2. Query Twap price
* Query twap price can be done by either submitting a message, or periodically
* If failed to query TWAP, exponential backoff will kick in. The query retry will happen after [1, 2, 4, 8, 16] epochs.
* If exponential backoff reachs 4, the connection will be considered OUTDATED.
* If success, the connection will be considered UPDATED, reset exponential backoff