package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

// TODO:  not use anymore, will remove this in v2.0.0
// SetOsmosisExchangeRate set osmosis exchange rate (osmosis to native token)
func (k Keeper) SetOsmosisExchangeRate(ctx sdk.Context, osmosisExchangeRate sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := osmosisExchangeRate.Marshal()
	store.Set(types.OsmosisTwapExchangeRate, bz)
}

// GetOsmosisExchangeRate get osmosis exchange rate (osmosis to native token)
// TODO:  not use anymore, will remove this in v2.0.0
func (k Keeper) GetOsmosisExchangeRate(ctx sdk.Context) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.OsmosisTwapExchangeRate)
	if bz == nil {
		return sdk.ZeroDec(), sdkerrors.Wrapf(types.ErrInvalidExchangeRate, "Osmosis does not have exchange rate data")
	}

	fmt.Println(string(bz))
	var osmosisExchangeRate sdk.Dec
	if err := osmosisExchangeRate.Unmarshal(bz); err != nil {
		panic(err)
	}

	return osmosisExchangeRate, nil
}

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
