package keeper

import (
<<<<<<< HEAD
	sdkerrors "cosmossdk.io/errors"
=======
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
>>>>>>> d2b5f20 (migrate from frozen to more generic host chain fee abs connection status (#156))
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/fee-abstraction/v4/x/feeabs/types"
)

<<<<<<< HEAD
func (keeper Keeper) HasHostZoneConfig(ctx sdk.Context, ibcDenom string) bool {
	store := ctx.KVStore(keeper.storeKey)
	key := types.GetKeyHostZoneConfig(ibcDenom)
=======
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
>>>>>>> d2b5f20 (migrate from frozen to more generic host chain fee abs connection status (#156))
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

<<<<<<< HEAD
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
		return sdkerrors.Wrapf(types.ErrHostZoneConfigNotFound, err.Error())
	}
	hostChainConfig.Frozen = false
	err = keeper.SetHostZoneConfig(ctx, ibcDenom, hostChainConfig)
=======
func (k Keeper) SetStateHostZoneByIBCDenom(ctx sdk.Context, ibcDenom string, state types.HostChainFeeAbsStatus) error {
	hostChainConfig, found := k.GetHostZoneConfig(ctx, ibcDenom)
	if !found {
		return types.ErrHostZoneConfigNotFound
	}
	hostChainConfig.Status = state
	err := k.SetHostZoneConfig(ctx, hostChainConfig)
>>>>>>> d2b5f20 (migrate from frozen to more generic host chain fee abs connection status (#156))
	if err != nil {
		return err
	}

	return nil
}
