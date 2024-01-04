package keeper

import (
	"fmt"
	"time"

	proto "github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
)

// HasEpochInfo return true if has epoch info
func (k Keeper) HasEpochInfo(ctx sdk.Context, identifier string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(append(types.KeyPrefixEpoch, []byte(identifier)...))
}

// GetEpochInfo returns epoch info by identifier.
func (k Keeper) GetEpochInfo(ctx sdk.Context, identifier string) types.EpochInfo {
	epoch := types.EpochInfo{}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(append(types.KeyPrefixEpoch, []byte(identifier)...))
	if b == nil {
		return epoch
	}
	err := proto.Unmarshal(b, &epoch)
	if err != nil {
		panic(err)
	}
	return epoch
}

// AddEpochInfo adds a new epoch info. Will return an error if the epoch fails validation,
// or re-uses an existing identifier.
// This method also sets the start time if left unset, and sets the epoch start height.
func (k Keeper) AddEpochInfo(ctx sdk.Context, epoch types.EpochInfo) error {
	err := epoch.Validate()
	if err != nil {
		return err
	}
	// Check if identifier already exists
	if (k.GetEpochInfo(ctx, epoch.Identifier) != types.EpochInfo{}) {
		return fmt.Errorf("epoch with identifier %s already exists", epoch.Identifier)
	}

	// Initialize empty and default epoch values
	if epoch.StartTime.Equal(time.Time{}) {
		epoch.StartTime = ctx.BlockTime()
	}
	epoch.CurrentEpochStartHeight = ctx.BlockHeight()
	k.setEpochInfo(ctx, epoch)
	return nil
}

// setEpochInfo set epoch info.
func (k Keeper) setEpochInfo(ctx sdk.Context, epoch types.EpochInfo) {
	store := ctx.KVStore(k.storeKey)
	value, err := proto.Marshal(&epoch)
	if err != nil {
		panic(err)
	}
	store.Set(append(types.KeyPrefixEpoch, []byte(epoch.Identifier)...), value)
}

// IterateEpochInfo iterate through epochs.
func (k Keeper) IterateEpochInfo(ctx sdk.Context, fn func(index int64, epochInfo types.EpochInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixEpoch)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		epoch := types.EpochInfo{}
		err := proto.Unmarshal(iterator.Value(), &epoch)
		if err != nil {
			panic(err)
		}
		stop := fn(i, epoch)

		if stop {
			break
		}
		i++
	}
}

// AllEpochInfos iterate through epochs to return all epochs info.
func (k Keeper) AllEpochInfos(ctx sdk.Context) []types.EpochInfo {
	epochs := []types.EpochInfo{}
	k.IterateEpochInfo(ctx, func(index int64, epochInfo types.EpochInfo) (stop bool) {
		epochs = append(epochs, epochInfo)
		return false
	})
	return epochs
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string) {
	switch epochIdentifier {
	case types.DefaultQueryEpochIdentifier:
		k.Logger(ctx).Info("Epoch interchain query TWAP")
		k.executeAllHostChainTWAPQuery(ctx)
	case types.DefaultSwapEpochIdentifier:
		k.Logger(ctx).Info("Epoch cross chain swap")
		k.executeAllHostChainSwap(ctx)
	default:
		k.Logger(ctx).Error(fmt.Sprintf("Unknown epoch %s", epochIdentifier))
	}
}
