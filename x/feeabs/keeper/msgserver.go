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

func (k Keeper) SendQuerySpotPrice(goCtx context.Context, msg *types.MsgSendQuerySpotPrice) (*types.MsgSendQuerySpotPriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}
	err = k.handleOsmosisIbcQuery(ctx)
	if err != nil {
		return nil, err
	}

	return &types.MsgSendQuerySpotPriceResponse{}, nil
}
