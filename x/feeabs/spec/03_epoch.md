# Epoch

The fee abstraction levegage the Osmosis `epoch` which is used to schedule the Inter-Blockchain Communication (IBC) send packet requests. These requests are for `RequestTwapData` and `SwapIBCToken`.

The `RequestTwapData` packet is used to request Time-Weighted Average Price (TWAP) data from Osmosis network.

The `SwapIBCToken` packet is for a feature of the fee abstraction module which allows for the transfer of IBC fees to the Osmosis cross-swap contract. The module account will then receive the native token associated with the fee..

Both these packets are scheduled by the fee abstraction module in accordance with the Osmosis epoch. This allows for efficient and timely transfer of data and tokens.