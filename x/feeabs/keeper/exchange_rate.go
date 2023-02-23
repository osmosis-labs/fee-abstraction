package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
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

func (k Keeper) SetTwapRate(ctx sdk.Context, ibcDenom string, osmosisTWAPExchangeRate sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := osmosisTWAPExchangeRate.Marshal()
	key := types.GetKeyTwapExchangeRate(ibcDenom)
	store.Set(key, bz)
}
