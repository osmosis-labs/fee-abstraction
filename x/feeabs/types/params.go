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
	KeyOsmosisIbcDenom                 = []byte("osmosisibcdenom")
	KeyOsmosisIbcConnectionId          = []byte("osmosisibcconnectionid")
	KeyOsmosisQueryContract            = []byte("osmosisquerycontract")
	KeyOsmosisExchangeRateUpdatePeriod = []byte("osmosisexchangerateupdateperiod")
	KeyAccumulatedOsmosisFeeSwapPeriod = []byte("accumulatedosmosisfeeswapperiod")

	_ paramtypes.ParamSet = &Params{}
)

// ParamTable for lockup module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyOsmosisIbcDenom, &p.OsmosisIbcDenom, validateOsmosisIbcDenom),
		paramtypes.NewParamSetPair(KeyOsmosisIbcConnectionId, &p.OsmosisIbcConnectionId, validateIbcConnectionId),
		paramtypes.NewParamSetPair(KeyOsmosisQueryContract, &p.OsmosisQueryContract, validateOsmosisQueryContract),
		paramtypes.NewParamSetPair(KeyOsmosisExchangeRateUpdatePeriod, &p.OsmosisExchangeRateUpdatePeriod, noOp),
		paramtypes.NewParamSetPair(KeyAccumulatedOsmosisFeeSwapPeriod, &p.AccumulatedOsmosisFeeSwapPeriod, noOp),
	}
}

// Validate also validates params info.
func (p Params) Validate() error {
	err := validateOsmosisIbcDenom(p.OsmosisIbcDenom)
	if err != nil {
		return fmt.Errorf("invalid ibc denom %s", err)
	}

	err = validateIbcConnectionId(p.OsmosisIbcConnectionId)
	if err != nil {
		return fmt.Errorf("invalid connection id %s", err)
	}

	err = validateOsmosisQueryContract(p.OsmosisQueryContract)
	if err != nil {
		return fmt.Errorf("invalid query contract %s", err)
	}

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

	// if strings.HasPrefix(denom, "ibc/") {
	// 	return fmt.Errorf("osmosis ibc denom doesn't have ibc prefix")
	// }

	return nil
}

func validateIbcConnectionId(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// if !strings.HasPrefix(connectionId, "connection-") {
	// 	return fmt.Errorf("wrong connection id format")
	// }

	return nil
}

func validateOsmosisQueryContract(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
