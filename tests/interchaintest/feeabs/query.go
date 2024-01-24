package feeabs

import (
	"context"
	"encoding/json"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
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

func QueryModuleAccountBalances(c *cosmos.CosmosChain, ctx context.Context) (*QueryFeeabsModuleBalacesResponse, error) {
	tn := getFullNode(c)
	cmd := []string{"feeabs", "module-balances"}
	stdout, _, err := tn.ExecQuery(ctx, cmd...)
	if err != nil {
		return &QueryFeeabsModuleBalacesResponse{}, err
	}

	var feeabsModule QueryFeeabsModuleBalacesResponse
	err = json.Unmarshal(stdout, &feeabsModule)
	if err != nil {
		return &QueryFeeabsModuleBalacesResponse{}, err
	}

	return &feeabsModule, nil
}
