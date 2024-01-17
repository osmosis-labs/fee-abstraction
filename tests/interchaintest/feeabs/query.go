package feeabs

import (
	"context"
	"encoding/json"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
)

type HostChainFeeAbsConfigResponse struct {
	HostChainConfig HostChainFeeAbsConfig `json:"host_chain_config"`
}

type HostChainFeeAbsConfig struct {
	IbcDenom                string `json:"ibc_denom"`
	OsmosisPoolTokenDenomIn string `json:"osmosis_pool_token_denom_in"`
	PoolId                  string `json:"pool_id"`
	Frozen                  bool   `json:"frozen"`
}

func QueryFeeabsHostZoneConfigWithDenom(c *cosmos.CosmosChain, ctx context.Context, denom string) (*HostChainFeeAbsConfigResponse, error) {
	cmd := []string{"feeabs", "host-chain-config", denom}
	stdout, _, err := c.ExecQuery(ctx, cmd)
	if err != nil {
		return &HostChainFeeAbsConfigResponse{}, err
	}

	var hostZoneConfig HostChainFeeAbsConfigResponse
	err = json.Unmarshal(stdout, &hostZoneConfig)
	if err != nil {
		return &HostChainFeeAbsConfigResponse{}, err
	}

	return &hostZoneConfig, nil
}
