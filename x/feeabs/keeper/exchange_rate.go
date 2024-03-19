package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

// GetTwapRate return Twap Price of ibcDenom
func (k Keeper) GetTwapRate(ctx sdk.Context, ibcDenom string) (sdkmath.LegacyDec, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyTwapExchangeRate(ibcDenom)
	bz := store.Get(key)
	if bz == nil {
		return sdkmath.LegacyDec{}, sdkerrors.Wrapf(types.ErrInvalidExchangeRate, "Osmosis does not have exchange rate data")
	}

	var osmosisExchangeRate sdkmath.LegacyDec
	if err := osmosisExchangeRate.Unmarshal(bz); err != nil {
		panic(err)
	}

	return osmosisExchangeRate, nil
}

func (k Keeper) SetTwapRate(ctx sdk.Context, ibcDenom string, osmosisTWAPExchangeRate sdkmath.LegacyDec) {
	store := ctx.KVStore(k.storeKey)
	bz, err := osmosisTWAPExchangeRate.Marshal()
	if err != nil {
		panic(err)
	}
	key := types.GetKeyTwapExchangeRate(ibcDenom)
	store.Set(key, bz)
}
