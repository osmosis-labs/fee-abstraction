package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// OsmosisSpotPrice return spot price of pair Osmo/nativeToken
func (q Querier) OsmosisArithmeticTwap(goCtx context.Context, req *types.QueryOsmosisArithmeticTwapRequest) (*types.QueryOsmosisArithmeticTwapResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	twapRate, err := q.GetTwapRate(ctx, req.IbcDenom)
	if err != nil {
		return nil, err
	}

	// TODO: move to use TWAP response
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
	}, nil
}

func (q Querier) HostChainConfig(goCtx context.Context, req *types.QueryHostChainConfigRequest) (*types.QueryHostChainConfigRespone, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	hostChainConfig, err := q.GetHostZoneConfig(ctx, req.IbcDenom)
	if err != nil {
		return nil, err
	}

	return &types.QueryHostChainConfigRespone{
		HostChainConfig: hostChainConfig,
	}, nil
}
