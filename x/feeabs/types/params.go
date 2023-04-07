package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Feeabs params default values .
const (
	DefaultOsmosisQueryTwapPath = "/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow"
	DefaultChainName            = "feeappd-t1"
	DefaultContractAddress      = ""
)

// Parameter keys store keys.
var (
	KeyOsmosisQueryTwapPath = []byte("osmosisquerytwappath")
	KeyNativeIbcedInOsmosis = []byte("nativeibcedinosmosis")
	KeyChainName            = []byte("chainname")

	_ paramtypes.ParamSet = &Params{}
)

// ParamTable for lockup module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyOsmosisQueryTwapPath, &p.OsmosisQueryTwapPath, validateOsmosisQueryTwapPath),
		paramtypes.NewParamSetPair(KeyNativeIbcedInOsmosis, &p.NativeIbcedInOsmosis, validateNativeIbcedInOsmosis),
		paramtypes.NewParamSetPair(KeyChainName, &p.ChainName, validateChainName),
	}
}

// Validate also validates params info.
func (p Params) Validate() error {

	if err := validateOsmosisQueryTwapPath(p.OsmosisQueryTwapPath); err != nil {
		return err
	}
	if err := validateNativeIbcedInOsmosis(p.NativeIbcedInOsmosis); err != nil {
		return err
	}

	return nil
}

func validateOsmosisQueryTwapPath(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type OsmosisQueryTwapPath: %T", i)
	}

	return nil
}

func validateNativeIbcedInOsmosis(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type NativeIbcedInOsmosis: %T", i)
	}

	return nil
}

func validateChainName(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type ChainName: %T", i)
	}

	return nil
}
