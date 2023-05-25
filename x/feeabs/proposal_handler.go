package feeabs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	v1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/notional-labs/fee-abstraction/v4/x/feeabs/keeper"

	cli "github.com/notional-labs/fee-abstraction/v4/x/feeabs/client/cli"
	"github.com/notional-labs/fee-abstraction/v4/x/feeabs/types"
)

var (
	UpdateAddHostZoneClientProposalHandler    = govclient.NewProposalHandler(cli.NewCmdSubmitAddHostZoneProposal)
	UpdateDeleteHostZoneClientProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitDeleteHostZoneProposal)
	UpdateSetHostZoneClientProposalHandler    = govclient.NewProposalHandler(cli.NewCmdSubmitSetHostZoneProposal)
)

// NewHostZoneProposal defines the add host zone proposal handler
func NewHostZoneProposal(k keeper.Keeper) v1beta1types.Handler {
	return func(ctx sdk.Context, content v1beta1types.Content) error {
		switch c := content.(type) {
		case *types.AddHostZoneProposal:
			return k.AddHostZoneProposal(ctx, c)
		case *types.DeleteHostZoneProposal:
			return k.DeleteHostZoneProposal(ctx, c)
		case *types.SetHostZoneProposal:
			return k.SetHostZoneProposal(ctx, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ibc proposal content type: %T", c)
		}
	}
}
