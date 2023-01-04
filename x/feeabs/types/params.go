package types

import (
	time "time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Feeabs params default values .
const (
	// After pass, ISO 8601 format for when swap in end period
	DefaultSwapPeriod time.Duration = time.Minute * 100

	// After pass, ISO 8601 format for when they can no longer burn EXP
	DefaultQueryPeriod time.Duration = time.Minute * 1

	// Contract address in Osmosis .
	DefaultContractAddress string = ""
)

// Parameter keys store keys.
var (
	KeyAllowedToken = []byte("allowed_token")
	KeySwapPeriod   = []byte("swap_period")
	KeyQueryPeriod  = []byte("query_period")

	_ paramtypes.ParamSet = &Params{}
)

// ParamTable for lockup module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAllowedToken, &p.AllowedToken, validateToken),
		paramtypes.NewParamSetPair(KeySwapPeriod, &p.SwapPeriod, validatePeriod),
		paramtypes.NewParamSetPair(KeyQueryPeriod, &p.SwapPeriod, validatePeriod),
	}
}

func validateToken(i interface{}) error {
	return nil
}
func validatePeriod(i interface{}) error {
	return nil
}
