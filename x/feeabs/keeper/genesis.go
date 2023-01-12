package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

// InitGenesis initializes the incentives module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// set Params
	k.SetParams(ctx, *genState.Params)
	// for IBC
	k.SetPort(ctx, genState.PortId)
	// Only try to bind to port if it is not already bound, since we may already own
	// port capability from capability InitGenesis
	if !k.IsBound(ctx, genState.PortId) {
		// module binds to the port on InitChain
		// and claims the returned capability
		err := k.BindPort(ctx, genState.PortId)
		if err != nil {
			panic("could not claim port capability: " + err.Error())
		}
	}

	for _, epoch := range genState.Epochs {
		err := k.AddEpochInfo(ctx, epoch)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the x/incentives module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	return &types.GenesisState{
		Params: &params,
		Epochs: k.AllEpochInfos(ctx),
		PortId: k.GetPort(ctx),
	}
}
