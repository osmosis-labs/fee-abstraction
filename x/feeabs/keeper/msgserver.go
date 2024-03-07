package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
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

func (k Keeper) SwapCrossChain(goCtx context.Context, msg *types.MsgSwapCrossChain) (*types.MsgSwapCrossChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}

	hostChainConfig, found := k.GetHostZoneConfig(ctx, msg.IbcDenom)
	if !found {
		return nil, types.ErrHostZoneConfigNotFound
	}

	if hostChainConfig.Status == types.HostChainFeeAbsStatus_FROZEN {
		return nil, types.ErrHostZoneFrozen
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
