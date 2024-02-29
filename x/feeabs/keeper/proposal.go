package keeper

import (
	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
)

func (k Keeper) AddHostZoneProposal(ctx sdk.Context, p *types.AddHostZoneProposal) error {
	if k.HasHostZoneConfig(ctx, p.HostChainConfig.IbcDenom) {
		return errors.Wrapf(types.ErrDuplicateHostZoneConfig, "duplicate host ibc denom")
	}

	if k.HasHostZoneConfigByOsmosisDenom(ctx, p.HostChainConfig.OsmosisPoolTokenDenomIn) {
		return errors.Wrapf(types.ErrDuplicateHostZoneConfig, "duplicate osmosis ibc denom")
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
		return errors.Wrapf(types.ErrHostZoneConfigNotFound, "host ibc denom not found: %s", p.HostChainConfig.IbcDenom)
	}

	// delete old host zone config
	if err := k.DeleteHostZoneConfig(ctx, p.HostChainConfig.IbcDenom); err != nil {
		return err
	}

	// set new host zone config
	if err := k.SetHostZoneConfig(ctx, *p.HostChainConfig); err != nil {
		return err
	}

	return nil
}
