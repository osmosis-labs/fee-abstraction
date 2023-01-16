package keeper

import (
	"context"
	"fmt"

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
func (q Querier) OsmosisSpotPrice(goCtx context.Context, req *types.QueryOsmosisSpotPriceRequest) (*types.QueryOsmosisSpotPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	spotPrice, err := q.GetOsmosisExchangeRate(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryOsmosisSpotPriceResponse{
		BaseAsset:  "osmo", // TODO: Currently hard code this value. Need to change to params then
		QuoteAsset: q.sk.BondDenom(ctx),
		SpotPrice:  spotPrice,
	}, nil
}

// FeeabsModuleBalances return total balances of feeabs module
func (q Querier) FeeabsModuleBalances(goCtx context.Context, req *types.QueryFeeabsModuleBalacesRequest) (*types.QueryFeeabsModuleBalacesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	moduleAddress := q.GetModuleAddress()
	fmt.Println("==================")
	fmt.Println(moduleAddress.String())
	fmt.Println("==================")

	moduleBalances := q.bk.GetAllBalances(ctx, moduleAddress)

	return &types.QueryFeeabsModuleBalacesResponse{
		Balances: moduleBalances,
	}, nil
}
