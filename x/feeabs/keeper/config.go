package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/fee-abstraction/v4/x/feeabs/types"
)

func (keeper Keeper) HasHostZoneConfig(ctx sdk.Context, ibcDenom string) bool {
	store := ctx.KVStore(keeper.storeKey)
	key := types.GetKeyHostZoneConfigByFeeabsIBCDenom(ibcDenom)
	return store.Has(key)
}

func (keeper Keeper) GetHostZoneConfig(ctx sdk.Context, ibcDenom string) (types.HostChainFeeAbsConfig, bool) {
	store := ctx.KVStore(keeper.storeKey)
	key := types.GetKeyHostZoneConfigByFeeabsIBCDenom(ibcDenom)

	var chainConfig types.HostChainFeeAbsConfig
	bz := store.Get(key)
	if bz == nil {
		return types.HostChainFeeAbsConfig{}, false
	}

	keeper.cdc.MustUnmarshal(bz, &chainConfig)

	return chainConfig, true
}

func (keeper Keeper) GetHostZoneConfigByOsmosisTokenDenom(ctx sdk.Context, osmosisIbcDenom string) (types.HostChainFeeAbsConfig, bool) {
	store := ctx.KVStore(keeper.storeKey)
	key := types.GetKeyHostZoneConfigByOsmosisIBCDenom(osmosisIbcDenom)

	var chainConfig types.HostChainFeeAbsConfig
	bz := store.Get(key)
	if bz == nil {
		return types.HostChainFeeAbsConfig{}, false
	}

	keeper.cdc.MustUnmarshal(bz, &chainConfig)

	return chainConfig, true
}

func (keeper Keeper) SetHostZoneConfig(ctx sdk.Context, chainConfig types.HostChainFeeAbsConfig) error {
	store := ctx.KVStore(keeper.storeKey)
	key := types.GetKeyHostZoneConfigByFeeabsIBCDenom(chainConfig.IbcDenom)

	bz, err := keeper.cdc.Marshal(&chainConfig)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	key = types.GetKeyHostZoneConfigByOsmosisIBCDenom(chainConfig.OsmosisPoolTokenDenomIn)
	store.Set(key, bz)

	return nil
}

func (keeper Keeper) DeleteHostZoneConfig(ctx sdk.Context, ibcDenom string) error {
	hostZoneConfig, _ := keeper.GetHostZoneConfig(ctx, ibcDenom)
	store := ctx.KVStore(keeper.storeKey)

	key := types.GetKeyHostZoneConfigByFeeabsIBCDenom(ibcDenom)
	store.Delete(key)

	key = types.GetKeyHostZoneConfigByOsmosisIBCDenom(hostZoneConfig.OsmosisPoolTokenDenomIn)
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
	return sdk.KVStorePrefixIterator(store, types.KeyHostChainChainConfigByFeeAbs)
}

// IterateHostZone iterates over the hostzone .
func (keeper Keeper) IterateHostZone(ctx sdk.Context, cb func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyHostChainChainConfigByFeeAbs)

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
	hostChainConfig, found := keeper.GetHostZoneConfig(ctx, ibcDenom)
	if !found {
		return types.ErrHostZoneConfigNotFound
	}
	hostChainConfig.Frozen = true
	err := keeper.SetHostZoneConfig(ctx, hostChainConfig)
	if err != nil {
		return err
	}

	return nil
}

func (keeper Keeper) UnFrozenHostZoneByIBCDenom(ctx sdk.Context, ibcDenom string) error {
	hostChainConfig, found := keeper.GetHostZoneConfig(ctx, ibcDenom)
	if !found {
		return types.ErrHostZoneConfigNotFound
	}
	hostChainConfig.Frozen = false
	err := keeper.SetHostZoneConfig(ctx, hostChainConfig)
	if err != nil {
		return err
	}

	return nil
}
