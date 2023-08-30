package keeper

import (
	sdkerrors "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
)

// GetTwapRate return Twap Price of ibcDenom
func (k Keeper) GetTwapRate(ctx sdk.Context, ibcDenom string) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyTwapExchangeRate(ibcDenom)
	bz := store.Get(key)
	if bz == nil {
		return sdk.ZeroDec(), sdkerrors.Wrapf(types.ErrInvalidExchangeRate, "Osmosis does not have exchange rate data")
	}

	var osmosisExchangeRate sdk.Dec
	if err := osmosisExchangeRate.Unmarshal(bz); err != nil {
		panic(err)
	}

	return osmosisExchangeRate, nil
}

// SetTwapRate set twap rate to state
func (k Keeper) SetTwapRate(ctx sdk.Context, ibcDenom string, osmosisTWAPExchangeRate sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := osmosisTWAPExchangeRate.Marshal()
	key := types.GetKeyTwapExchangeRate(ibcDenom)
	store.Set(key, bz)
}

// SetSendingPacketInfo store the sending icq packet hostchain
func (k Keeper) SetSendingPacketInfo(ctx sdk.Context, sequence uint64, channel string, hostZoneConfig types.HostChainFeeAbsConfig) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyIcqTwapSequence)
	key := types.GetKeyIcqTwapSequence(sequence, channel)

	bz := k.cdc.MustMarshal(&hostZoneConfig)
	prefixStore.Set(key, bz)
}

// GetAndRemoveSendingPacketInfo get and remove the sending icq packet hostchain
func (k Keeper) GetAndRemoveSendingPacketInfo(ctx sdk.Context, sequence uint64, channel string) types.HostChainFeeAbsConfig {
	var hostZoneConfig types.HostChainFeeAbsConfig
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyIcqTwapSequence)
	key := types.GetKeyIcqTwapSequence(sequence, channel)

	bz := prefixStore.Get(key)
	k.cdc.MustUnmarshal(bz, &hostZoneConfig)

	prefixStore.Delete(key)
	return hostZoneConfig
}
