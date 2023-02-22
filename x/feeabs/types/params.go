package types

import (
	"fmt"
	time "time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Feeabs params default values .
const (
	DefaultSwapPeriod time.Duration = time.Minute * 100

	DefaultQueryPeriod time.Duration = time.Minute * 1

	DefaultContractAddress string = ""
)

// Parameter keys store keys.
var (
	KeyOsmosisExchangeRateUpdatePeriod = []byte("osmosisexchangerateupdateperiod")
	KeyAccumulatedOsmosisFeeSwapPeriod = []byte("accumulatedosmosisfeeswapperiod")
	KeyNativeIbcedInOsmosis            = []byte("nativeibcedinosmosis")

	_ paramtypes.ParamSet = &Params{}
)

// ParamTable for lockup module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyOsmosisExchangeRateUpdatePeriod, &p.OsmosisExchangeRateUpdatePeriod, noOp),
		paramtypes.NewParamSetPair(KeyAccumulatedOsmosisFeeSwapPeriod, &p.AccumulatedOsmosisFeeSwapPeriod, noOp),
		paramtypes.NewParamSetPair(KeyNativeIbcedInOsmosis, &p.NativeIbcedInOsmosis, noOp),
	}
}

// Validate also validates params info.
func (p Params) Validate() error {

	if p.OsmosisExchangeRateUpdatePeriod == 0 {
		return fmt.Errorf("invalid zero period")
	}
	if p.AccumulatedOsmosisFeeSwapPeriod == 0 {
		return fmt.Errorf("invalid zero period")
	}

	return nil
}

func noOp(i interface{}) error {
	return nil
}

func validateOsmosisIbcDenom(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateChannelID(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateOsmosisQueryContract(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validatePoolID(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateActive(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
