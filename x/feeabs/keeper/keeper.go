package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct{}

// need to implement
func (k Keeper) GetModuleAddress() sdk.AccAddress {
	return sdk.AccAddress{}
}

// need to implement
func (k Keeper) CalculateNativeFromIBCCoin(ibcCoin sdk.Coins) (coins sdk.Coins, err error) {
	err = k.verifyIBCCoin(ibcCoin)
	if err != nil {
		return sdk.Coins{}, nil
	}
	return coins, nil
}

// TODO : need to implement
// return err if IBC token isn't in allowed_list
func (k Keeper) verifyIBCCoin(ibcCoin sdk.Coins) error {
	return nil
}
