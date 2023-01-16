package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

// SetOsmosisExchangeRate set osmosis exchange rate (osmosis to native token)
func (k Keeper) SetOsmosisExchangeRate(ctx sdk.Context, osmosisExchangeRate sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := osmosisExchangeRate.Marshal()
	store.Set(types.OsmosisExchangeRate, bz)
}

// GetOsmosisExchangeRate get osmosis exchange rate (osmosis to native token)
func (k Keeper) GetOsmosisExchangeRate(ctx sdk.Context) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.OsmosisExchangeRate)
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
