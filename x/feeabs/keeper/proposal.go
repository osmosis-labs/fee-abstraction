package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"

	errorsmod "cosmossdk.io/errors"
)

func (k Keeper) AddHostZoneProposal(ctx sdk.Context, p *types.AddHostZoneProposal) error {
	// Check if duplicate host zone
	if k.HasHostZoneConfig(ctx, p.HostChainConfig.IbcDenom) {
		return errorsmod.Wrapf(types.ErrDuplicateHostZoneConfig, "duplicate IBC denom")
	}
	if k.HasHostZoneConfigByOsmosisTokenDenom(ctx, p.HostChainConfig.OsmosisPoolTokenDenomIn) {
		return errorsmod.Wrapf(types.ErrDuplicateHostZoneConfig, "duplicate Osmosis's IBC denom")
	}

	err := k.SetHostZoneConfig(ctx, *p.HostChainConfig)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) DeleteHostZoneProposal(ctx sdk.Context, p *types.DeleteHostZoneProposal) error {
	_, found := k.GetHostZoneConfig(ctx, p.IbcDenom)
	if !found {
		return types.ErrHostZoneConfigNotFound
	}

	err := k.DeleteHostZoneConfig(ctx, p.IbcDenom)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) SetHostZoneProposal(ctx sdk.Context, p *types.SetHostZoneProposal) error {
	_, found := k.GetHostZoneConfig(ctx, p.HostChainConfig.IbcDenom)
	if !found {
		return types.ErrHostZoneConfigNotFound
	}

	// Delete all hostzone
	err := k.DeleteHostZoneConfig(ctx, p.HostChainConfig.IbcDenom)
	if err != nil {
		return err
	}

	// set new hostzone
	err = k.SetHostZoneConfig(ctx, *p.HostChainConfig)
	if err != nil {
		return err
	}

	return nil
}
