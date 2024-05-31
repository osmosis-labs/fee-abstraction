package feeabs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"

	feeabstypes "github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

func QueryHostZoneConfigWithDenom(c *cosmos.CosmosChain, ctx context.Context, denom string) (*HostChainFeeAbsConfigResponse, error) {
	tn := c.GetNode()
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

func QueryAllHostZoneConfig(c *cosmos.CosmosChain, ctx context.Context) (*AllQueryHostChainConfigResponse, error) {
	tn := c.GetNode()
	cmd := []string{"feeabs", "all-host-chain-config"}
	stdout, _, err := tn.ExecQuery(ctx, cmd...)
	if err != nil {
		return &AllQueryHostChainConfigResponse{}, err
	}

	var hostZoneConfig AllQueryHostChainConfigResponse
	err = json.Unmarshal(stdout, &hostZoneConfig)
	if err != nil {
		return &AllQueryHostChainConfigResponse{}, err
	}

	return &hostZoneConfig, nil
}

func QueryModuleAccountBalances(c *cosmos.CosmosChain, ctx context.Context) (*feeabstypes.QueryFeeabsModuleBalacesResponse, error) {
	tn := c.GetNode()
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

// QueryOsmosisArithmeticTwap queries the arithmetic twap of ibc denom stored in fee abstraction module
func QueryOsmosisArithmeticTwap(c *cosmos.CosmosChain, ctx context.Context, ibcDenom string) (*feeabstypes.QueryOsmosisArithmeticTwapResponse, error) {
	node := c.GetNode()
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

// QueryOsmosisArithmeticTwapOsmosis queries the arithmetic twap of a pool on osmosis chain
func QueryOsmosisArithmeticTwapOsmosis(c *cosmos.CosmosChain, ctx context.Context, poolID, ibcDenom string) (*feeabstypes.QueryOsmosisArithmeticTwapResponse, error) {
	node := c.GetNode()
	currentEpoch := time.Now().Unix()

	cmd := []string{"twap", "arithmetic", poolID, ibcDenom, fmt.Sprintf("%d", currentEpoch-20), fmt.Sprintf("%d", currentEpoch-10)}
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
