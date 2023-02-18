package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

func (k Keeper) AddHostZoneProposal(ctx sdk.Context, p *types.AddHostZoneProposal) error {
	_, err := k.GetHostZoneConfig(ctx, p.HostChainConfig.IbcDenom)
	if err == nil {
		return types.ErrDuplicateHostZoneConfig
	}

	err = k.SetHostZoneConfig(ctx, p.HostChainConfig.IbcDenom, *p.HostChainConfig)
	if err != nil {
		return err
	}

	return nil
}
