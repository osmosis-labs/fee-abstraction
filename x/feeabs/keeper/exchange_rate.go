package keeper

import (
	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
)

// GetTwapRate return Twap Price of ibcDenom
func (k Keeper) GetTwapRate(ctx sdk.Context, ibcDenom string) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyTwapExchangeRate(ibcDenom)
	bz := store.Get(key)
	if bz == nil {
		return sdk.Dec{}, sdkerrors.Wrapf(types.ErrInvalidExchangeRate, "Osmosis does not have exchange rate data")
	}

	var osmosisExchangeRate sdk.Dec
	if err := osmosisExchangeRate.Unmarshal(bz); err != nil {
		panic(err)
	}

	return osmosisExchangeRate, nil
}

func (k Keeper) SetTwapRate(ctx sdk.Context, ibcDenom string, osmosisTWAPExchangeRate sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz, err := osmosisTWAPExchangeRate.Marshal()
	if err != nil {
		panic(err)
	}
	key := types.GetKeyTwapExchangeRate(ibcDenom)
	store.Set(key, bz)
}
