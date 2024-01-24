package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/feeabstraction keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// OsmosisArithmeticTwap return spot price of pair Osmo/nativeToken
func (q Querier) OsmosisArithmeticTwap(goCtx context.Context, req *types.QueryOsmosisArithmeticTwapRequest) (*types.QueryOsmosisArithmeticTwapResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	twapRate, err := q.GetTwapRate(ctx, req.IbcDenom)
	if err != nil {
		return nil, err
	}

	return &types.QueryOsmosisArithmeticTwapResponse{
		ArithmeticTwap: twapRate,
	}, nil
}

// FeeabsModuleBalances return total balances of feeabs module
func (q Querier) FeeabsModuleBalances(goCtx context.Context, req *types.QueryFeeabsModuleBalacesRequest) (*types.QueryFeeabsModuleBalacesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	moduleAddress := q.GetFeeAbsModuleAddress()
	moduleBalances := q.bk.GetAllBalances(ctx, moduleAddress)

	return &types.QueryFeeabsModuleBalacesResponse{
		Balances: moduleBalances,
		Address:  moduleAddress.String(),
	}, nil
}

func (q Querier) HostChainConfig(goCtx context.Context, req *types.QueryHostChainConfigRequest) (*types.QueryHostChainConfigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	hostChainConfig, found := q.GetHostZoneConfig(ctx, req.IbcDenom)
	if !found {
		return nil, types.ErrHostZoneConfigNotFound
	}

	return &types.QueryHostChainConfigResponse{
		HostChainConfig: hostChainConfig,
	}, nil
}

func (q Querier) AllHostChainConfig(goCtx context.Context, req *types.AllQueryHostChainConfigRequest) (*types.AllQueryHostChainConfigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	allHostChainConfig, err := q.GetAllHostZoneConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &types.AllQueryHostChainConfigResponse{
		AllHostChainConfig: allHostChainConfig,
	}, nil
}
