package app

import (
	"testing"

	"cosmossdk.io/log"
	"github.com/CosmWasm/wasmd/app/params"
	dbm "github.com/cosmos/cosmos-db"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
)

// MakeEncodingConfig creates a new EncodingConfig with all modules registered. For testing only
func MakeEncodingConfig(t testing.TB) params.EncodingConfig {
	t.Helper()
	// we "pre"-instantiate the application for getting the injected/configured encoding configuration
	// note, this is not necessary when using app wiring, as depinject can be directly used (see root_v2.go)
	tempApp := NewFeeApp(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
	return makeEncodingConfig(tempApp)
}

func makeEncodingConfig(tempApp *FeeApp) params.EncodingConfig {
	encodingConfig := params.EncodingConfig{
		InterfaceRegistry: tempApp.InterfaceRegistry(),
		Codec:             tempApp.AppCodec(),
		TxConfig:          tempApp.TxConfig(),
		Amino:             tempApp.LegacyAmino(),
	}
	return encodingConfig
}
