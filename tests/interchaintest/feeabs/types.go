package feeabs

type HostChainFeeAbsConfigResponse struct {
	HostChainConfig HostChainFeeAbsConfig `json:"host_chain_config"`
}

type AllQueryHostChainConfigResponse struct {
	AllHostChainConfig []HostChainFeeAbsConfig `json:"all_host_chain_config"`
}

const (
	HostChainFeeAbsStatus_UPDATED  string = "UPDATED"
	HostChainFeeAbsStatus_OUTDATED string = "OUTDATED"
	HostChainFeeAbsStatus_FROZEN   string = "FROZEN"
)

type HostChainFeeAbsConfig struct {
	// ibc token is allowed to be used as fee token
	IbcDenom string `protobuf:"bytes,1,opt,name=ibc_denom,json=ibcDenom,proto3" json:"ibc_denom,omitempty" yaml:"allowed_token"`
	// token_in in cross_chain swap contract.
	OsmosisPoolTokenDenomIn string `protobuf:"bytes,2,opt,name=osmosis_pool_token_denom_in,json=osmosisPoolTokenDenomIn,proto3" json:"osmosis_pool_token_denom_in,omitempty"`
	// pool id
	PoolId string `protobuf:"varint,3,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	// Host chain fee abstraction connection status
	Status string `protobuf:"varint,4,opt,name=status,proto3,enum=feeabstraction.feeabs.v1beta1.HostChainFeeAbsStatus" json:"status,omitempty"`
}
