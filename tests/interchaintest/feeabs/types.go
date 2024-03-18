package feeabs

type HostChainFeeAbsConfigResponse struct {
	HostChainConfig HostChainFeeAbsConfig `json:"host_chain_config"`
}

const (
	HostChainFeeAbsStatus_UPDATED  string = "UPDATED"
	HostChainFeeAbsStatus_OUTDATED string = "OUTDATED"
	HostChainFeeAbsStatus_FROZEN   string = "FROZEN"
)

type HostChainFeeAbsConfig struct {
	// ibc token is allowed to be used as fee token
	IbcDenom string `json:"ibc_denom,omitempty"`
	// token_in in cross_chain swap contract.
	OsmosisPoolTokenDenomIn string `json:"osmosis_pool_token_denom_in,omitempty"`
	// pool id
	PoolId string `json:"pool_id,omitempty"`
	// Host chain fee abstraction connection status
	Status        string `json:"status,omitempty"`
	MinSwapAmount uint64 `json:"min_swap_amount,omitempty"`
}
