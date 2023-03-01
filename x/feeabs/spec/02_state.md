# State 

## OsmosisTwapExchangeRate

The exchange rate of an ibc denom to Osmosis: `0x01<ibc_denom_bytes> -> sdk.Dec`

When we send the QueryArithmeticTwapToNowRequest to the Osmosis contract via IBC, the contract will send an acknowledgement with price data to the fee abstraction chain. The OsmosisTwapExchangeRate will then be updated based on this value.
This exchange rate is then used to calculate transaction fees in the appropriate IBC denom. By updating the exchange rate based on the most recent price data, we can ensure that transaction fees accurately reflect the current market conditions on Osmosis.

It's important to note that the exchange rate will fluctuate over time, as it is based on the time-weighted average price (TWAP) of the IBC denom on Osmosis. This means that the exchange rate will reflect the average price of the IBC denom over a certain time period, rather than an instantaneous price.

## HostChainChainConfig

The host chain config for an ibc denom
- KeyHostChainChainConfig: `0x03<ibc_denom_bytes> -> HostChainFeeAbsConfig`

```go
type HostChainFeeAbsConfig struct {
	// ibc token is allowed to be used as fee token
	IbcDenom string `protobuf:"bytes,1,opt,name=ibc_denom,json=ibcDenom,proto3" json:"ibc_denom,omitempty" yaml:"allowed_token"`
	// token_in in cross_chain swap contract.
	OsmosisPoolTokenDenomIn string `protobuf:"bytes,2,opt,name=osmosis_pool_token_denom_in,json=osmosisPoolTokenDenomIn,proto3" json:"osmosis_pool_token_denom_in,omitempty"`
	// TODO: middleware address in hostchain, can we refator this logic ?
	MiddlewareAddress string `protobuf:"bytes,3,opt,name=middleware_address,json=middlewareAddress,proto3" json:"middleware_address,omitempty"`
	// transfer channel from customer_chain -> host chain
	IbcTransferChannel string `protobuf:"bytes,4,opt,name=ibc_transfer_channel,json=ibcTransferChannel,proto3" json:"ibc_transfer_channel,omitempty"`
	// transfer channel from host chain -> osmosis
	HostZoneIbcTransferChannel string `protobuf:"bytes,5,opt,name=host_zone_ibc_transfer_channel,json=hostZoneIbcTransferChannel,proto3" json:"host_zone_ibc_transfer_channel,omitempty"`
	// crosschain-swap contract address
	CrosschainSwapAddress string `protobuf:"bytes,6,opt,name=crosschain_swap_address,json=crosschainSwapAddress,proto3" json:"crosschain_swap_address,omitempty"`
	// pool id
	PoolId uint64 `protobuf:"varint,7,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	// Active
	IsOsmosis bool `protobuf:"varint,8,opt,name=is_osmosis,json=isOsmosis,proto3" json:"is_osmosis,omitempty"`
	// Frozen
	Frozen bool `protobuf:"varint,9,opt,name=frozen,proto3" json:"frozen,omitempty"`
	// Query channel
	OsmosisQueryChannel string `protobuf:"bytes,10,opt,name=osmosis_query_channel,json=osmosisQueryChannel,proto3" json:"osmosis_query_channel,omitempty"`
}
```


rpc : http://168.119.91.22:2241
api:  http://168.119.91.22:1318