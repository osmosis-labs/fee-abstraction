package feeabs

import (
	"context"
	"encoding/json"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"

	feeabstypes "github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

func QueryHostZoneConfigWithDenom(c *cosmos.CosmosChain, ctx context.Context, denom string) (*HostChainFeeAbsConfigResponse, error) {
	tn := getFullNode(c)
	cmd := []string{"feeabs", "host-chain-config", denom}
	stdout, _, err := tn.ExecQuery(ctx, cmd...)
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

func QueryHostZoneConfig(c *cosmos.CosmosChain, ctx context.Context) (*HostChainFeeAbsConfigResponse, error) {
	tn := getFullNode(c)
	cmd := []string{"feeabs", "all-host-chain-config"}
	stdout, _, err := tn.ExecQuery(ctx, cmd...)
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

func QueryModuleAccountBalances(c *cosmos.CosmosChain, ctx context.Context) (*feeabstypes.QueryFeeabsModuleBalacesResponse, error) {
	tn := getFullNode(c)
	cmd := []string{"feeabs", "module-balances"}
	stdout, _, err := tn.ExecQuery(ctx, cmd...)
	if err != nil {
		return &feeabstypes.QueryFeeabsModuleBalacesResponse{}, err
	}

	var response feeabstypes.QueryFeeabsModuleBalacesResponse
	if err = json.Unmarshal(stdout, &response); err != nil {
		return &feeabstypes.QueryFeeabsModuleBalacesResponse{}, err
	}

	return &response, nil
}

func QueryOsmosisArithmeticTwap(c *cosmos.CosmosChain, ctx context.Context, ibcDenom string) (*feeabstypes.QueryOsmosisArithmeticTwapResponse, error) {
	node := getFullNode(c)
	cmd := []string{"feeabs", "osmo-arithmetic-twap", ibcDenom}
	stdout, _, err := node.ExecQuery(ctx, cmd...)
	if err != nil {
		return &feeabstypes.QueryOsmosisArithmeticTwapResponse{}, err
	}

	var response feeabstypes.QueryOsmosisArithmeticTwapResponse
	if err = json.Unmarshal(stdout, &response); err != nil {
		return &feeabstypes.QueryOsmosisArithmeticTwapResponse{}, err
	}
	return &response, nil
}
