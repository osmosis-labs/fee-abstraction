package keeper

import (
	"context"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
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
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	_, err = k.HandleOsmosisIbcQuery(ctx)
	if err != nil {
		return nil, err
	}

	return &types.MsgSendQueryIbcDenomTWAPResponse{}, nil
}

func (k Keeper) SwapCrossChain(goCtx context.Context, msg *types.MsgSwapCrossChain) (*types.MsgSwapCrossChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(msg.Sender)
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

	if hostChainConfig.Status == types.HostChainFeeAbsStatus_OUTDATED {
		return nil, types.ErrHostZoneOutdated
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
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	err = k.bk.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgFundFeeAbsModuleAccountResponse{}, nil
}

func (k msgServer) UpdateParams(ctx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	sdkContext := sdk.UnwrapSDKContext(ctx)
	if k.GetAuthority() != req.Authority {
		return nil, errors.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	if err := req.Params.Validate(); err != nil {
		return nil, err
	}

	k.SetParams(sdkContext, req.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) AddHostZone(ctx context.Context, req *types.MsgAddHostZone) (*types.MsgAddHostZoneResponse, error) {
	sdkContext := sdk.UnwrapSDKContext(ctx)
	if k.GetAuthority() != req.Authority {
		return nil, errors.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}
	if k.HasHostZoneConfig(sdkContext, req.HostChainConfig.IbcDenom) {
		return nil, errors.Wrapf(types.ErrDuplicateHostZoneConfig, "duplicate host ibc denom")
	}

	if err := k.SetHostZoneConfig(sdkContext, req.HostChainConfig); err != nil {
		return nil, err
	}

	return &types.MsgAddHostZoneResponse{}, nil
}

func (k msgServer) RemoveHostZone(ctx context.Context, req *types.MsgRemoveHostZone) (*types.MsgRemoveHostZoneResponse, error) {
	sdkContext := sdk.UnwrapSDKContext(ctx)
	if k.GetAuthority() != req.Authority {
		return nil, errors.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	return &types.MsgRemoveHostZoneResponse{}, k.DeleteHostZoneConfig(sdkContext, req.IbcDenom)
}

func (k msgServer) UpdateHostZone(ctx context.Context, req *types.MsgUpdateHostZone) (*types.MsgUpdateHostZoneResponse, error) {
	sdkContext := sdk.UnwrapSDKContext(ctx)
	if k.GetAuthority() != req.Authority {
		return nil, errors.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	if !k.HasHostZoneConfig(sdkContext, req.HostChainConfig.IbcDenom) {
		return nil, errors.Wrapf(types.ErrHostZoneConfigNotFound, "host zone config not found")
	}

	if err := k.SetHostZoneConfig(sdkContext, req.HostChainConfig); err != nil {
		return nil, err
	}

	return &types.MsgUpdateHostZoneResponse{}, nil
}
