package types

import (
	"errors"
	"fmt"
)

// DefaultGenesis returns the incentive module's default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: &Params{
			OsmosisIbcDenom: "ibc/",
			OsmosisIbcConnectionId: "",
			OsmosisQueryContract: "",
			OsmosisExchangeRateUpdatePeriod: DefaultQueryPeriod,
			AccumulatedOsmosisFeeSwapPeriod: DefaultSwapPeriod,
		},
		Epochs: []EpochInfo{NewGenesisEpochInfo("swap", DefaultQueryPeriod)},
	}
}

// Validate performs basic genesis state validation, returning an error upon any failure.
func (gs GenesisState) Validate() error {

	//Validate params
	err := gs.Params.Validate()
	if err != nil {
		return fmt.Errorf("invalid params", err)
	}

	// Validate epochs genesis
	for _, epoch := range(gs.Epochs) {
		err := epoch.Validate()
		if err != nil {
			return errors.New("invalid epoch")
		}
	}
	return nil
}