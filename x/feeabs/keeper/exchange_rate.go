package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

// SetOsmosisExchangeRate set osmosis exchange rate (osmosis to native token)
func (k Keeper) SetOsmosisExchangeRate(ctx sdk.Context, osmosisExchangeRate string) {
	store := ctx.KVStore(k.storeKey)
	spotPriceData := types.SpotPriceData{
		SpotPrice: osmosisExchangeRate,
	}
	bz, err := k.cdc.Marshal(&spotPriceData)
	// TODO: handler logic here, refactor that trash
	fmt.Println("================")
	fmt.Println(err)
	fmt.Println("================")

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
	var spotPrice types.SpotPriceData
	k.cdc.Unmarshal(bz, &spotPrice)
	spotPriceDec, err := sdk.NewDecFromStr(spotPrice.SpotPrice)
	if err != nil {
		return sdk.Dec{}, sdkerrors.New("ibc ack data umarshal", 1, "error when NewDecFromStr")
	}

	return spotPriceDec, nil
}
