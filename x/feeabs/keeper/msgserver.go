package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/notional-labs/fee-abstraction/v2/x/feeabs/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{
		Keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

// Need to remove this
func (k Keeper) SendQueryIbcDenomTWAP(goCtx context.Context, msg *types.MsgSendQueryIbcDenomTWAP) (*types.MsgSendQueryIbcDenomTWAPResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}

	startTime := ctx.BlockTime().Add(-msg.Duration)
	k.Logger(ctx).Info(fmt.Sprintf("Start time: %v", startTime.Unix()))

	params := k.GetParams(ctx)

	var reqs []types.QueryArithmeticTwapToNowRequest
	k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
		req := types.NewQueryArithmeticTwapToNowRequest(
			hostZoneConfig.PoolId,
			params.NativeIbcedInOsmosis,
			hostZoneConfig.OsmosisPoolTokenDenomIn,
			startTime,
		)
		reqs = append(reqs, req)
		return false
	})
	err = k.SendOsmosisQueryRequest(ctx, reqs, types.IBCPortID, params.IbcQueryIcqChannel)
	if err != nil {
		return nil, err
	}

	return &types.MsgSendQueryIbcDenomTWAPResponse{}, nil
}

// Need to remove this
func (k Keeper) SwapCrossChain(goCtx context.Context, msg *types.MsgSwapCrossChain) (*types.MsgSwapCrossChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	hostChainConfig, err := k.GetHostZoneConfig(ctx, msg.IbcDenom)
	if err != nil {
		return &types.MsgSwapCrossChainResponse{}, nil
	}
	_, err = sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}

	err = k.transferOsmosisCrosschainSwap(ctx, hostChainConfig)

	if err != nil {
		return nil, err
	}

	return &types.MsgSwapCrossChainResponse{}, nil
}

func (k Keeper) FundFeeAbsModuleAccount(
	goCtx context.Context,
	msg *types.MsgFundFeeAbsModuleAccount,
) (*types.MsgFundFeeAbsModuleAccountResponse, error) {
	// Unwrap context
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Check sender address
	sender, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}

	err = k.bk.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgFundFeeAbsModuleAccountResponse{}, nil
}
