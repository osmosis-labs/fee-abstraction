package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

func (keeper Keeper) HasHostZoneConfig(ctx sdk.Context, ibcDenom string) bool {
	store := ctx.KVStore(keeper.storeKey)
	key := types.GetKeyHostZoneConfig(ibcDenom)
	return store.Has(key)
}

func (keeper Keeper) GetHostZoneConfig(ctx sdk.Context, ibcDenom string) (chainConfig types.HostChainFeeAbsConfig, err error) {
	store := ctx.KVStore(keeper.storeKey)
	key := types.GetKeyHostZoneConfig(ibcDenom)

	bz := store.Get(key)
	err = keeper.cdc.Unmarshal(bz, &chainConfig)

	if err != nil {
		return types.HostChainFeeAbsConfig{}, err
	}

	return
}

func (keeper Keeper) SetHostZoneConfig(ctx sdk.Context, ibcDenom string, chainConfig types.HostChainFeeAbsConfig) error {
	store := ctx.KVStore(keeper.storeKey)
	key := types.GetKeyHostZoneConfig(ibcDenom)

	bz, err := keeper.cdc.Marshal(&chainConfig)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	return nil
}

func (keeper Keeper) DeleteHostZoneConfig(ctx sdk.Context, ibcDenom string) error {
	store := ctx.KVStore(keeper.storeKey)
	key := types.GetKeyHostZoneConfig(ibcDenom)
	store.Delete(key)
	return nil
}

// use iterator
func (keeper Keeper) GetAllHostZoneConfig(ctx sdk.Context) (allChainConfigs []types.HostChainFeeAbsConfig, err error) {
	keeper.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
		allChainConfigs = append(allChainConfigs, hostZoneConfig)
		return false
	})

	return allChainConfigs, nil
}

func (keeper Keeper) IteratorHostZone(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, types.KeyHostChainChainConfig)
}

// IterateHostZone iterates over the hostzone .
// TODO: write test for this .
func (keeper Keeper) IterateHostZone(ctx sdk.Context, cb func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyHostChainChainConfig)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var hostZoneConfig types.HostChainFeeAbsConfig
		keeper.cdc.MustUnmarshal(iterator.Value(), &hostZoneConfig)
		if cb(hostZoneConfig) {
			break
		}
	}
}

func (keeper Keeper) FrozenHostZoneByIBCDenom(ctx sdk.Context, ibcDenom string) error {
	hostChainConfig, err := keeper.GetHostZoneConfig(ctx, ibcDenom)
	if err != nil {
		// TODO: registry the error here
		return sdkerrors.Wrapf(types.ErrHostZoneConfigNotFound, err.Error())
	}
	hostChainConfig.Frozen = true
	err = keeper.SetHostZoneConfig(ctx, ibcDenom, hostChainConfig)
	if err != nil {
		return err
	}

	return nil
}

func (keeper Keeper) UnFrozenHostZoneByIBCDenom(ctx sdk.Context, ibcDenom string) error {
	hostChainConfig, err := keeper.GetHostZoneConfig(ctx, ibcDenom)
	if err != nil {
		// TODO: registry the error here
		return sdkerrors.Wrapf(types.ErrHostZoneConfigNotFound, err.Error())
	}
	hostChainConfig.Frozen = false
	err = keeper.SetHostZoneConfig(ctx, ibcDenom, hostChainConfig)
	if err != nil {
		return err
	}

	return nil
}
