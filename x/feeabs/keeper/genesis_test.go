package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	apphelpers "github.com/notional-labs/feeabstraction/v1/app/helpers"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

var now = time.Now().UTC()

var testGenesis = types.GenesisState{
	Params: types.Params{
		OsmosisExchangeRateUpdatePeriod: types.DefaultQueryPeriod,
		AccumulatedOsmosisFeeSwapPeriod: types.DefaultSwapPeriod,
		NativeIbcedInOsmosis:            "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878",
	},
	Epochs: []types.EpochInfo{
		types.NewGenesisEpochInfo("query", types.DefaultQueryPeriod),
		types.NewGenesisEpochInfo("swap", types.DefaultSwapPeriod),
	},
	PortId: types.IBCPortID,
}

func TestInitGenesis(t *testing.T) {
	app := apphelpers.Setup(t, false, 1)
	ctx := apphelpers.NewContextForApp(*app)

	ctx = ctx.WithBlockHeight(1)
	ctx = ctx.WithBlockTime(now)
	genesis := testGenesis

	params := app.FeeabsKeeper.GetParams(ctx)
	require.Equal(t, params, genesis.Params)

	epochs := app.FeeabsKeeper.AllEpochInfos(ctx)
	require.Equal(t, epochs, genesis.Epochs)

	portid := app.FeeabsKeeper.GetPort(ctx)
	require.Equal(t, portid, genesis.PortId)
}

func TestExportGenesis(t *testing.T) {
	app := apphelpers.Setup(t, false, 1)
	ctx := apphelpers.NewContextForApp(*app)
	ctx = ctx.WithBlockHeight(1)
	genesis := app.FeeabsKeeper.ExportGenesis(ctx)
	require.Len(t, genesis.Epochs, 2)

	expectedEpochs := types.DefaultGenesis().Epochs
	require.Equal(t, expectedEpochs, genesis.Epochs)
}
