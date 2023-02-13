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
// func (k Keeper) SendQuerySpotPrice(goCtx context.Context, msg *types.MsgSendQuerySpotPrice) (*types.MsgSendQuerySpotPriceResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)

// 	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
// 	if err != nil {
// 		return nil, err
// 	}
// 	hostChainConfig, err := k.GetHostZoneConfig(ctx, chainID)
// 	if err != nil {
// 		return &types.MsgSendQuerySpotPriceResponse{}, nil
// 	}

// 	err = k.handleOsmosisIbcQuery(ctx, hostChainConfig)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &types.MsgSendQuerySpotPriceResponse{}, nil
// }

func (k Keeper) SendQuerySpotPrice(goCtx context.Context, msg *types.MsgSendQuerySpotPrice) (*types.MsgSendQuerySpotPriceResponse, error) {
	return &types.MsgSendQuerySpotPriceResponse{}, nil
}

// Need to remove this
// func (k Keeper) SwapCrossChain(goCtx context.Context, msg *types.MsgSwapCrossChain, chainID string) (*types.MsgSwapCrossChainResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)
// 	hostChainConfig, err := k.GetHostZoneConfig(ctx, chainID)
// 	if err != nil {
// 		return &types.MsgSwapCrossChainResponse{}, nil
// 	}
// 	_, err = sdk.AccAddressFromBech32(msg.FromAddress)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = k.transferIBCTokenToOsmosisContract(ctx, hostChainConfig)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &types.MsgSwapCrossChainResponse{}, nil
// }

func (k Keeper) SwapCrossChain(goCtx context.Context, msg *types.MsgSwapCrossChain) (*types.MsgSwapCrossChainResponse, error) {
	return &types.MsgSwapCrossChainResponse{}, nil
}

func (k Keeper) InterchainQueryBalances(goCtx context.Context, msg *types.MsgInterchainQueryBalances) (*types.MsgInterchainQueryBalancesRespone, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("IBC InterchainQueryBalances")

	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}
	k.Logger(ctx).Error("IBC InterchainQueryBalances handleInterchainQuery")
	err = k.handleInterchainQuery(ctx, msg.QueryAddress)
	if err != nil {
		return nil, err
	}

	return &types.MsgInterchainQueryBalancesRespone{}, nil
}
