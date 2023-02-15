package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyHostChainChainConfig)

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
