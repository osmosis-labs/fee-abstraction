// Reference: https://github.com/cosmos/gaia/blob/main/x/globalfee/ante/fee_utils.go
package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// splitCoinsByDenoms returns the given coins split in two whether
// their demon is or isn't found in the given denom map.
func splitCoinsByDenoms(feeCoins sdk.Coins, denomMap map[string]struct{}) (sdk.Coins, sdk.Coins) {
	feeCoinsNonZeroDenom, feeCoinsZeroDenom := sdk.Coins{}, sdk.Coins{}

	for _, fc := range feeCoins {
		_, found := denomMap[fc.Denom]
		if found {
			feeCoinsZeroDenom = append(feeCoinsZeroDenom, fc)
		} else {
			feeCoinsNonZeroDenom = append(feeCoinsNonZeroDenom, fc)
		}
	}

	return feeCoinsNonZeroDenom.Sort(), feeCoinsZeroDenom.Sort()
}

// getNonZeroFees returns the given fees nonzero coins
// and a map storing the zero coins's denoms
func getNonZeroFees(fees sdk.Coins) (sdk.Coins, map[string]struct{}) {
	requiredFeesNonZero := sdk.Coins{}
	requiredFeesZeroDenom := map[string]struct{}{}

	for _, gf := range fees {
		if gf.IsZero() {
			requiredFeesZeroDenom[gf.Denom] = struct{}{}
		} else {
			requiredFeesNonZero = append(requiredFeesNonZero, gf)
		}
	}

	return requiredFeesNonZero.Sort(), requiredFeesZeroDenom
}
