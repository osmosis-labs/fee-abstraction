package types

import fmt "fmt"

// DefaultGenesis returns the incentive module's default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: &Params{
			OsmosisIbcDenom:                 "ibc/",
			OsmosisQueryChannel:             "",
			OsmosisTransferChannel:          "",
			OsmosisSwapContract:             "",
			OsmosisQueryContract:            "",
			OsmosisExchangeRateUpdatePeriod: DefaultQueryPeriod,
			AccumulatedOsmosisFeeSwapPeriod: DefaultSwapPeriod,
			PoolId:                          1,
			Active:                          false,
		},
		Epochs: []EpochInfo{NewGenesisEpochInfo("query", DefaultQueryPeriod), NewGenesisEpochInfo("swap", DefaultSwapPeriod)},
		PortId: IBCPortID,
	}
}

// Validate performs basic genesis state validation, returning an error upon any failure.
func (gs GenesisState) Validate() error {
	//Validate params
	err := gs.Params.Validate()
	if err != nil {
		return fmt.Errorf("invalid params %s", err)
	}

	// Validate epochs genesis
	for _, epoch := range gs.Epochs {
		err := epoch.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}
