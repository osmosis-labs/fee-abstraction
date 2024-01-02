package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
)

func (k Keeper) AddHostZoneProposal(ctx sdk.Context, p *types.AddHostZoneProposal) error {
	if k.HasHostZoneConfig(ctx, p.HostChainConfig.IbcDenom) {
		return types.ErrDuplicateHostZoneConfig
	}

	if err := k.SetHostZoneConfig(ctx, *p.HostChainConfig); err != nil {
		return err
	}

	return nil
}

func (k Keeper) DeleteHostZoneProposal(ctx sdk.Context, p *types.DeleteHostZoneProposal) error {
	return k.DeleteHostZoneConfig(ctx, p.IbcDenom)
}

func (k Keeper) SetHostZoneProposal(ctx sdk.Context, p *types.SetHostZoneProposal) error {
	if !k.HasHostZoneConfig(ctx, p.HostChainConfig.IbcDenom) {
		return types.ErrHostZoneConfigNotFound
	}

	if err := k.SetHostZoneConfig(ctx, *p.HostChainConfig); err != nil {
		return err
	}

	return nil
}
