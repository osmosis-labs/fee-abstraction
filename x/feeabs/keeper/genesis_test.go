package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	feeapp "github.com/notional-labs/feeabstraction/v1/app"
	apphelpers "github.com/notional-labs/feeabstraction/v1/app/helpers"
	"github.com/notional-labs/feeabstraction/v1/x/feeabs"
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
		types.NewGenesisEpochInfo("query1", types.DefaultQueryPeriod),
	},
	PortId: types.IBCPortID,
}

func TestMarshalUnmarshalGenesis(t *testing.T) {
	app := apphelpers.Setup(t, false, 1)
	ctx := apphelpers.NewContextForApp(*app)
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	encodingConfig := feeapp.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	am := feeabs.NewAppModule(appCodec, app.FeeabsKeeper)
	genesis := testGenesis
	app.FeeabsKeeper.InitGenesis(ctx, genesis)

	genesisExported := am.ExportGenesis(ctx, appCodec)
	assert.NotPanics(t, func() {
		app := apphelpers.Setup(t, false, 1)
		ctx := apphelpers.NewContextForApp(*app)
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := feeabs.NewAppModule(appCodec, app.FeeabsKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}

func TestInitGenesis(t *testing.T) {
	app := apphelpers.Setup(t, false, 1)
	ctx := apphelpers.NewContextForApp(*app)

	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	app.FeeabsKeeper.InitGenesis(ctx, genesis)

	params := app.FeeabsKeeper.GetParams(ctx)
	require.Equal(t, params, genesis.Params)

	epochs := app.FeeabsKeeper.AllEpochInfos(ctx)
	require.Equal(t, epochs, genesis.Epochs)

	portid := app.FeeabsKeeper.GetPort(ctx)
	require.Equal(t, portid, genesis.PortId)
}

// func TestExportGenesis(t *testing.T) {
// 	app := simapp.Setup(false)
// 	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
// 	ctx = ctx.WithBlockTime(now.Add(time.Second))
// 	genesis := testGenesis
// 	app.SuperfluidKeeper.InitGenesis(ctx, genesis)

// 	asset := types.SuperfluidAsset{
// 		Denom:     "gamm/pool/2",
// 		AssetType: types.SuperfluidAssetTypeLPShare,
// 	}
// 	app.SuperfluidKeeper.SetSuperfluidAsset(ctx, asset)
// 	savedAsset := app.SuperfluidKeeper.GetSuperfluidAsset(ctx, "gamm/pool/2")
// 	require.Equal(t, savedAsset, asset)

// 	genesisExported := app.SuperfluidKeeper.ExportGenesis(ctx)
// 	require.Equal(t, genesisExported.Params, genesis.Params)
// 	require.Equal(t, genesisExported.SuperfluidAssets, append(genesis.SuperfluidAssets, asset))
// 	require.Equal(t, genesis.OsmoEquivalentMultipliers, genesis.OsmoEquivalentMultipliers)
// 	require.Equal(t, genesis.IntermediaryAccounts, genesis.IntermediaryAccounts)
// 	require.Equal(t, genesis.IntemediaryAccountConnections, genesis.IntemediaryAccountConnections)
// }
