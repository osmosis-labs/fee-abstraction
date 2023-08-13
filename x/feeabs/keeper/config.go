package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/fee-abstraction/v4/x/feeabs/types"
)

func (k Keeper) HasHostZoneConfig(ctx sdk.Context, ibcDenom string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyHostZoneConfig(ibcDenom)
	return store.Has(key)
}

func (k Keeper) GetHostZoneConfig(ctx sdk.Context, ibcDenom string) (chainConfig types.HostChainFeeAbsConfig, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyHostZoneConfig(ibcDenom)

	bz := store.Get(key)
	err = k.cdc.Unmarshal(bz, &chainConfig)

	if err != nil {
		return types.HostChainFeeAbsConfig{}, err
	}

	return chainConfig, nil
}

func (k Keeper) SetHostZoneConfig(ctx sdk.Context, ibcDenom string, chainConfig types.HostChainFeeAbsConfig) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyHostZoneConfig(ibcDenom)

	bz, err := k.cdc.Marshal(&chainConfig)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	return nil
}

func (k Keeper) DeleteHostZoneConfig(ctx sdk.Context, ibcDenom string) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyHostZoneConfig(ibcDenom)
	store.Delete(key)
	return nil
}

// use iterator
func (k Keeper) GetAllHostZoneConfig(ctx sdk.Context) (allChainConfigs []types.HostChainFeeAbsConfig, err error) {
	k.IterateHostZone(ctx, func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool) {
		allChainConfigs = append(allChainConfigs, hostZoneConfig)
		return false
	})

	return allChainConfigs, nil
}

func (k Keeper) IteratorHostZone(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, types.KeyHostChainChainConfig)
}

// IterateHostZone iterates over the hostzone .
func (k Keeper) IterateHostZone(ctx sdk.Context, cb func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyHostChainChainConfig)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var hostZoneConfig types.HostChainFeeAbsConfig
		k.cdc.MustUnmarshal(iterator.Value(), &hostZoneConfig)
		if cb(hostZoneConfig) {
			break
		}
	}
}

func (k Keeper) FrozenHostZoneByIBCDenom(ctx sdk.Context, ibcDenom string) error {
	hostChainConfig, err := k.GetHostZoneConfig(ctx, ibcDenom)
	if err != nil {
		// TODO: registry the error here
		return sdkerrors.Wrapf(types.ErrHostZoneConfigNotFound, err.Error())
	}
	hostChainConfig.Frozen = true
	err = k.SetHostZoneConfig(ctx, ibcDenom, hostChainConfig)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) UnFrozenHostZoneByIBCDenom(ctx sdk.Context, ibcDenom string) error {
	hostChainConfig, err := k.GetHostZoneConfig(ctx, ibcDenom)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrHostZoneConfigNotFound, err.Error())
	}
	hostChainConfig.Frozen = false
	err = k.SetHostZoneConfig(ctx, ibcDenom, hostChainConfig)
	if err != nil {
		return err
	}

	return nil
}
