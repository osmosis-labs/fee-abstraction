package keeper

import (
	"fmt"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

func (k Keeper) IncreaseBlockDelayToQuery(ctx sdk.Context, ibcDenom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.StoreExponentialBackoff)

	// must have host zone
	if !k.HasHostZoneConfig(ctx, ibcDenom) {
		panic("host zone config not found")
	}

	// must have query epoch info
	currentEpoch, exist := k.GetEpochInfo(ctx, types.DefaultQueryEpochIdentifier)
	if !exist {
		panic("epoch not found")
	}

	// get current exponential backoff
	currentJump := k.GetBlockDelayToQuery(ctx, ibcDenom).Jump
	nextJump := currentJump * 2
	if nextJump > types.ExponentialMaxJump {
		nextJump = types.ExponentialMaxJump
	}

	fmt.Println("currentJump", currentJump, "nextJump", nextJump, "currentEpoch.CurrentEpoch", currentEpoch.CurrentEpoch, "nextEpoch", currentEpoch.CurrentEpoch+nextJump)

	next := &types.ExponentialBackoff{
		Jump:        nextJump,
		FutureEpoch: currentEpoch.CurrentEpoch + nextJump,
	}

	store.Set([]byte(ibcDenom), k.cdc.MustMarshal(next))
}

func (k Keeper) ResetBlockDelayToQuery(ctx sdk.Context, ibcDenom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.StoreExponentialBackoff)

	// must have host zone
	if !k.HasHostZoneConfig(ctx, ibcDenom) {
		panic("host zone config not found")
	}

	// FutureEpoch = 0, current epoch will always be greater, thus always querying twap
	next := &types.ExponentialBackoff{
		Jump:        1,
		FutureEpoch: 0,
	}

	store.Set([]byte(ibcDenom), k.cdc.MustMarshal(next))
}

func (k Keeper) GetBlockDelayToQuery(ctx sdk.Context, ibcDenom string) types.ExponentialBackoff {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.StoreExponentialBackoff)
	bz := store.Get([]byte(ibcDenom))
	if bz == nil {
		return types.ExponentialBackoff{
			Jump:        1,
			FutureEpoch: 0,
		}
	}

	var next types.ExponentialBackoff
	k.cdc.MustUnmarshal(bz, &next)

	return next
}

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

	key := types.GetKeyHostZoneConfigByOsmosisIBCDenom(hostZoneConfig.OsmosisPoolTokenDenomIn)
	store.Delete(key)

	key = types.GetKeyHostZoneConfigByFeeabsIBCDenom(ibcDenom)
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
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyHostChainConfigByFeeAbs)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var hostZoneConfig types.HostChainFeeAbsConfig
		k.cdc.MustUnmarshal(iterator.Value(), &hostZoneConfig)
		if cb(hostZoneConfig) {
			break
		}
	}
}

func (k Keeper) SetStateHostZoneByIBCDenom(ctx sdk.Context, ibcDenom string, state types.HostChainFeeAbsStatus) error {
	hostChainConfig, found := k.GetHostZoneConfig(ctx, ibcDenom)
	if !found {
		return types.ErrHostZoneConfigNotFound
	}
	hostChainConfig.Status = state
	err := k.SetHostZoneConfig(ctx, hostChainConfig)
	if err != nil {
		return err
	}

	return nil
}
