package feeabs

import "github.com/cosmos/cosmos-sdk/types"

type HostChainFeeAbsConfigResponse struct {
	HostChainConfig HostChainFeeAbsConfig `json:"host_chain_config"`
}

type HostChainFeeAbsConfig struct {
	IbcDenom                string `json:"ibc_denom"`
	OsmosisPoolTokenDenomIn string `json:"osmosis_pool_token_denom_in"`
	PoolId                  string `json:"pool_id"`
	Frozen                  bool   `json:"frozen"`
}

type AddHostZoneProposalType struct {
	Title           string                `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description     string                `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	HostChainConfig HostChainFeeAbsConfig `protobuf:"bytes,3,opt,name=host_chain_config,json=hostChainConfig,proto3" json:"host_chain_config,omitempty"`
	Deposit         string                `json:"deposit"`
}

type QueryFeeabsModuleBalacesResponse struct {
	Balances types.Coins
	Address  string
}
