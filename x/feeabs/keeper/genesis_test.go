package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	app "github.com/osmosis-labs/fee-abstraction/v8/app"
	"github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

// var now = time.Now().UTC()

var defaultGenesis = types.DefaultGenesis()

func TestInitGenesis(t *testing.T) {
	app := app.Setup(t)
	ctx := app.NewContext(false)

	ctx = ctx.WithBlockHeight(1)
	genesis := defaultGenesis

	params := app.FeeabsKeeper.GetParams(ctx)
	require.Equal(t, params, genesis.Params)

	epochs := app.FeeabsKeeper.AllEpochInfos(ctx)
	require.Equal(t, epochs, genesis.Epochs)

	portid := app.FeeabsKeeper.GetPort(ctx)
	require.Equal(t, portid, genesis.PortId)
}

func TestExportGenesis(t *testing.T) {
	app := app.Setup(t)
	ctx := app.NewContext(false)

	ctx = ctx.WithBlockHeight(1)
	genesis := app.FeeabsKeeper.ExportGenesis(ctx)
	require.Len(t, genesis.Epochs, 2)

	expectedEpochs := types.DefaultGenesis().Epochs
	require.Equal(t, expectedEpochs, genesis.Epochs)
}
