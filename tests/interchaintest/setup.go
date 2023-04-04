package interchaintest

import (
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	feeabstype "github.com/notional-labs/fee-abstraction/v2/x/feeabs/types"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
)

var (
	FeeabsMainRepo = "ghcr.io/notional-labs/fee-abstraction"

	feeabsImage = ibc.DockerImage{
		Repository: "ghcr.io/notional-labs/fee-abstraction",
		Version:    "lastest",
		UidGid:     "1025:1025",
	}

	feeabsConfig = ibc.ChainConfig{
		Type:                "cosmos",
		Name:                "feeabs",
		ChainID:             "feeabs-2",
		Images:              []ibc.DockerImage{feeabsImage},
		Bin:                 "feeappd",
		Bech32Prefix:        "feeabs",
		Denom:               "stake",
		CoinType:            "118",
		GasPrices:           "0.0stake",
		GasAdjustment:       1.1,
		TrustingPeriod:      "112h",
		NoHostMount:         false,
		SkipGenTx:           false,
		PreGenesis:          nil,
		ModifyGenesis:       nil,
		ConfigFileOverrides: nil,
		EncodingConfig:      feeabsEncoding(),
	}

	pathFeeabsGaia      = "feeabs-gaia"
	genesisWalletAmount = int64(10_000_000)
)

// feeabsEncoding registers the feeabs specific module codecs so that the associated types and msgs
// will be supported when writing to the blocksdb sqlite database.
func feeabsEncoding() *simappparams.EncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	feeabstype.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}
