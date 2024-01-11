package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
)

func (k Keeper) HasHostZoneConfig(ctx sdk.Context, ibcDenom string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyHostZoneConfigByFeeabsIBCDenom(ibcDenom)
	return store.Has(key)
}

func (k Keeper) GetHostZoneConfig(ctx sdk.Context, ibcDenom string) (types.HostChainFeeAbsConfig, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyHostZoneConfigByFeeabsIBCDenom(ibcDenom)

	var chainConfig types.HostChainFeeAbsConfig
	bz := store.Get(key)
	if bz == nil {
		return types.HostChainFeeAbsConfig{}, false
	}

	k.cdc.MustUnmarshal(bz, &chainConfig)

	return chainConfig, true
}

func (k Keeper) GetHostZoneConfigByOsmosisTokenDenom(ctx sdk.Context, osmosisIbcDenom string) (types.HostChainFeeAbsConfig, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyHostZoneConfigByOsmosisIBCDenom(osmosisIbcDenom)

	var chainConfig types.HostChainFeeAbsConfig
	bz := store.Get(key)
	if bz == nil {
		return types.HostChainFeeAbsConfig{}, false
	}

	k.cdc.MustUnmarshal(bz, &chainConfig)

	return chainConfig, true
}

func (k Keeper) SetHostZoneConfig(ctx sdk.Context, chainConfig types.HostChainFeeAbsConfig) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyHostZoneConfigByFeeabsIBCDenom(chainConfig.IbcDenom)

	bz, err := k.cdc.Marshal(&chainConfig)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	key = types.GetKeyHostZoneConfigByOsmosisIBCDenom(chainConfig.OsmosisPoolTokenDenomIn)
	store.Set(key, bz)

	return nil
}

func (k Keeper) DeleteHostZoneConfig(ctx sdk.Context, ibcDenom string) error {
	hostZoneConfig, ok := k.GetHostZoneConfig(ctx, ibcDenom)
	if !ok {
		return types.ErrHostZoneConfigNotFound
	}
	store := ctx.KVStore(k.storeKey)

	key := types.GetKeyHostZoneConfigByFeeabsIBCDenom(ibcDenom)
	store.Delete(key)

	key = types.GetKeyHostZoneConfigByOsmosisIBCDenom(hostZoneConfig.OsmosisPoolTokenDenomIn)
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

// IterateHostZone iterates over the hostzone .
func (k Keeper) IterateHostZone(ctx sdk.Context, cb func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyHostChainConfigByFeeAbs)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var hostZoneConfig types.HostChainFeeAbsConfig
		k.cdc.MustUnmarshal(iterator.Value(), &hostZoneConfig)
		if cb(hostZoneConfig) {
			break
		}
	}
}

func (k Keeper) FreezeHostZoneByIBCDenom(ctx sdk.Context, ibcDenom string) error {
	hostChainConfig, found := k.GetHostZoneConfig(ctx, ibcDenom)
	if !found {
		return types.ErrHostZoneConfigNotFound
	}
	hostChainConfig.Frozen = true
	err := k.SetHostZoneConfig(ctx, hostChainConfig)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) UnFreezeHostZoneByIBCDenom(ctx sdk.Context, ibcDenom string) error {
	hostChainConfig, found := k.GetHostZoneConfig(ctx, ibcDenom)
	if !found {
		return types.ErrHostZoneConfigNotFound
	}
	hostChainConfig.Frozen = false
	err := k.SetHostZoneConfig(ctx, hostChainConfig)
	if err != nil {
		return err
	}

	return nil
}
