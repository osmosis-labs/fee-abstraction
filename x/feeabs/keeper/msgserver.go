package keeper

import (
	"context"
	"fmt"

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

	fmt.Println("=========allthings=================")
	fmt.Println(k.GetAllHostZoneConfig(ctx))
	fmt.Println("==========================")
	// k.RemoveHostZoneConfig(ctx, "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518")
	fmt.Println(k.GetHostZoneConfig(ctx, "ibc"))
	fmt.Println("==========================")
	k.handleOsmosisIbcQuery(ctx)
	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}

	return &types.MsgSendQueryIbcDenomTWAPResponse{}, nil
}

// Need to remove this
func (k Keeper) SwapCrossChain(goCtx context.Context, msg *types.MsgSwapCrossChain) (*types.MsgSwapCrossChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	hostChainConfig, err := k.GetHostZoneConfig(ctx, "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518")
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
