package keeper_test

import (
	"testing"
	"time"

	apphelpers "github.com/osmosis-labs/fee-abstraction/v7/app/helpers"
	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
	"github.com/stretchr/testify/require"
)

var now = time.Now().UTC()

var testGenesis = types.DefaultGenesis()

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
