package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/fee-abstraction/v4/x/feeabs/types"
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
	_, err = k.HandleOsmosisIbcQuery(ctx)
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

<<<<<<< HEAD
=======
	hostChainConfig, found := k.GetHostZoneConfig(ctx, msg.IbcDenom)
	if !found {
		return nil, types.ErrHostZoneConfigNotFound
	}

	if hostChainConfig.Status == types.HostChainFeeAbsStatus_FROZEN {
		return nil, types.ErrHostZoneFrozen
	}

	if hostChainConfig.Status == types.HostChainFeeAbsStatus_OUTDATED {
		return nil, types.ErrHostZoneOutdated
	}

>>>>>>> d2b5f20 (migrate from frozen to more generic host chain fee abs connection status (#156))
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
