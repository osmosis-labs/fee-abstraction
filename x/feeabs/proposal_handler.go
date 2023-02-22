package feeabs

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/keeper"

	cli "github.com/notional-labs/feeabstraction/v1/x/feeabs/client/cli"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

var (
	UpdateClientProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitAddHostZoneProposal, emptyRestHandler)
)

// NewAddHostZoneProposal defines the add host zone proposal handler
func NewAddHostZoneProposal(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.AddHostZoneProposal:
			return k.AddHostZoneProposal(ctx, c)
		// TODO : add remove host zone here.
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ibc proposal content type: %T", c)
		}
	}
}

// TODO : support this @Gnad @Ducnt.
func emptyRestHandler(client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: "unsupported",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "Legacy REST Routes are not supported")
		},
	}
}
