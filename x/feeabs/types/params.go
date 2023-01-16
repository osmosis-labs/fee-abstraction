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
	KeyNativeIbcDenom                  = []byte("nativeibcdenom")
	KeyOsmosisQueryChannel             = []byte("osmosisquerychannel")
	KeyOsmosisTransferChannel          = []byte("osmosistransferchannel")
	KeyOsmosisQueryContract            = []byte("osmosisquerycontract")
	KeyOsmosisSwapContract             = []byte("osmosisswapcontract")
	KeyOsmosisExchangeRateUpdatePeriod = []byte("osmosisexchangerateupdateperiod")
	KeyAccumulatedOsmosisFeeSwapPeriod = []byte("accumulatedosmosisfeeswapperiod")
	KeyPoolId                          = []byte("poolid")
	KeyActive                          = []byte("active")
	KeyAddress                         = []byte("address")

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
		paramtypes.NewParamSetPair(KeyNativeIbcDenom, &p.NativeIbcDenom, validateOsmosisIbcDenom),
		paramtypes.NewParamSetPair(KeyOsmosisQueryChannel, &p.OsmosisQueryChannel, validateChannelID),
		paramtypes.NewParamSetPair(KeyOsmosisTransferChannel, &p.OsmosisTransferChannel, validateChannelID),
		paramtypes.NewParamSetPair(KeyOsmosisQueryContract, &p.OsmosisQueryContract, validateOsmosisQueryContract),
		paramtypes.NewParamSetPair(KeyOsmosisSwapContract, &p.OsmosisSwapContract, validateOsmosisQueryContract),
		paramtypes.NewParamSetPair(KeyOsmosisExchangeRateUpdatePeriod, &p.OsmosisExchangeRateUpdatePeriod, noOp),
		paramtypes.NewParamSetPair(KeyAccumulatedOsmosisFeeSwapPeriod, &p.AccumulatedOsmosisFeeSwapPeriod, noOp),
		paramtypes.NewParamSetPair(KeyPoolId, &p.PoolId, validatePoolID),
		paramtypes.NewParamSetPair(KeyActive, &p.Active, validateActive),
	}
}

// Validate also validates params info.
func (p Params) Validate() error {
	err := validateOsmosisIbcDenom(p.OsmosisIbcDenom)
	if err != nil {
		return fmt.Errorf("invalid ibc denom %s", err)
	}

	err = validateChannelID(p.OsmosisQueryChannel)
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
