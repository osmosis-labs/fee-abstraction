package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

func (k Keeper) AddHostZoneProposal(ctx sdk.Context, p *types.AddHostZoneProposal) error {
	config, _ := k.GetHostZoneConfig(ctx, p.HostChainConfig.IbcDenom)
	if (config != types.HostChainFeeAbsConfig{}) {
		return types.ErrDuplicateHostZoneConfig
	}

	err := k.SetHostZoneConfig(ctx, p.HostChainConfig.IbcDenom, *p.HostChainConfig)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) DeleteHostZoneProposal(ctx sdk.Context, p *types.DeleteHostZoneProposal) error {
	_, err := k.GetHostZoneConfig(ctx, p.HostChainConfig.IbcDenom)
	if err == nil {
		return types.ErrHostZoneConfigNotFound
	}

	err = k.DeleteHostZoneConfig(ctx, p.HostChainConfig.IbcDenom)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) SetHostZoneProposal(ctx sdk.Context, p *types.SetHostZoneProposal) error {
	_, err := k.GetHostZoneConfig(ctx, p.HostChainConfig.IbcDenom)
	if err == nil {
		return types.ErrHostZoneConfigNotFound
	}

	err = k.SetHostZoneConfig(ctx, p.HostChainConfig.IbcDenom, *p.HostChainConfig)
	if err != nil {
		return err
	}

	return nil
}
