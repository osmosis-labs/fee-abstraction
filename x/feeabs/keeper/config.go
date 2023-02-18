package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

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

// use iterator
func (keeper Keeper) GetAllHostZoneConfig(ctx sdk.Context) (allChainConfigs []types.HostChainFeeAbsConfig, err error) {
	iterator := keeper.IteratorHostZone(ctx)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var chainConfig types.HostChainFeeAbsConfig
		err := keeper.cdc.Unmarshal(bz, &chainConfig)
		if err != nil {
			panic(err)
		}
		allChainConfigs = append(allChainConfigs, chainConfig)
	}

	return allChainConfigs, nil
}

func (keeper Keeper) IteratorHostZone(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, types.KeyHostChainChainConfig)
}

// IteraterHostZone iterates over the hostzone .
func (keeper Keeper) IteraterHostZone(ctx sdk.Context, cb func(hostZoneConfig types.HostChainFeeAbsConfig) (stop bool)) {
	iterator := keeper.IteratorHostZone(ctx)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		hostZoneConfig, err := keeper.GetHostZoneConfig(ctx, string(iterator.Key()))
		if err != nil {
			panic(err)
		}

		if cb(hostZoneConfig) {
			break
		}
	}
}

func (keeper Keeper) FronzenHostZoneByIBCDenom(ctx sdk.Context, ibcDenom string) error {
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
