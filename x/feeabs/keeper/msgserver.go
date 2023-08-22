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
