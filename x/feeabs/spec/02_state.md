# State 

## OsmosisTwapExchangeRate

The exchange rate of an ibc denom to Osmosis: `0x01<ibc_denom_bytes> -> sdk.Dec`

When we send the QueryArithmeticTwapToNowRequest to the Osmosis contract via IBC, the contract will send an acknowledgement with price data to the fee abstraction chain. The OsmosisTwapExchangeRate will then be updated based on this value.
This exchange rate is then used to calculate transaction fees in the appropriate IBC denom. By updating the exchange rate based on the most recent price data, we can ensure that transaction fees accurately reflect the current market conditions on Osmosis.

It's important to note that the exchange rate will fluctuate over time, as it is based on the time-weighted average price (TWAP) of the IBC denom on Osmosis. This means that the exchange rate will reflect the average price of the IBC denom over a certain time period, rather than an instantaneous price.
