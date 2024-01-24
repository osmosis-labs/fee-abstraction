package feeabs

type HostChainFeeAbsConfigResponse struct {
	HostChainConfig HostChainFeeAbsConfig `json:"host_chain_config"`
}

type HostChainFeeAbsConfig struct {
	IbcDenom                string `json:"ibc_denom"`
	OsmosisPoolTokenDenomIn string `json:"osmosis_pool_token_denom_in"`
	PoolId                  string `json:"pool_id"`
	Frozen                  bool   `json:"frozen"`
}
