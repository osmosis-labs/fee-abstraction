package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

// SetOsmosisExchangeRate set osmosis exchange rate (osmosis to native token)
func (k Keeper) SetOsmosisExchangeRate(ctx sdk.Context, osmosisExchangeRate sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: osmosisExchangeRate})
	store.Set(types.GetOsmosisExchangeRateKey(), bz)
}

// GetOsmosisExchangeRate get osmosis exchange rate (osmosis to native token)
func (k Keeper) GetOsmosisExchangeRate(ctx sdk.Context) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetOsmosisExchangeRateKey())
	if bz == nil {
		return sdk.ZeroDec(), sdkerrors.Wrapf(types.ErrInvalidExchangeRate, "Osmosis does not have exchange rate data")
	}
	var decProto sdk.DecProto
	k.cdc.MustUnmarshal(bz, &decProto)
	return decProto.Dec, nil
}
