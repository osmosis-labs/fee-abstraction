package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
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
	err = k.handleOsmosisIbcQuery(ctx)
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

	// TODO: don't use if/else logic
	if hostChainConfig.IsOsmosis {
		err = k.transferIBCTokenToOsmosisChainWithIBCHookMemo(ctx, hostChainConfig)
	} else {
		err = k.transferIBCTokenToHostChainWithMiddlewareMemo(ctx, hostChainConfig)
	}

	if err != nil {
		return nil, err
	}

	return &types.MsgSwapCrossChainResponse{}, nil
}
