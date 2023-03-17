
## HostChainChainConfig

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

## Configuring HostZoneConfig

In order to use Fee Abstraction, we need to add the HostZoneConfig as specified in the government proposals.
