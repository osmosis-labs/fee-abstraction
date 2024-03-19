package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	app "github.com/osmosis-labs/fee-abstraction/v8/app"
	feeabskeeper "github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/keeper"
	"github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

func createNHostZone(t *testing.T, keeper *feeabskeeper.Keeper, ctx sdk.Context, n int) []types.HostChainFeeAbsConfig {
	t.Helper()
	var expected []types.HostChainFeeAbsConfig
	expectedConfig := types.HostChainFeeAbsConfig{
		IbcDenom:                "ibc/123",
		OsmosisPoolTokenDenomIn: "ibc/456",
		PoolId:                  1,
		Status:                  types.HostChainFeeAbsStatus_UPDATED,
	}
	for i := 0; i < n; i++ {
		expected = append(expected, expectedConfig)
		err := keeper.SetHostZoneConfig(ctx, expectedConfig)
		require.NoError(t, err)
	}
	return expected
}

func TestHostZoneGet(t *testing.T) {
	app := app.Setup(t)
	ctx := app.NewContextLegacy(true, cmtproto.Header{Height: 1})

	expected := createNHostZone(t, &app.FeeabsKeeper, ctx, 1)
	for _, item := range expected {
		got, found := app.FeeabsKeeper.GetHostZoneConfig(ctx, item.IbcDenom)
		require.True(t, found)
		require.Equal(t, item, got)
	}
}

func TestHostZoneGetByOsmosisDenom(t *testing.T) {
	app := app.Setup(t)
	ctx := app.NewContextLegacy(true, cmtproto.Header{Height: 1})

	expected := createNHostZone(t, &app.FeeabsKeeper, ctx, 1)
	for _, item := range expected {
		got, found := app.FeeabsKeeper.GetHostZoneConfigByOsmosisTokenDenom(ctx, item.OsmosisPoolTokenDenomIn)
		require.True(t, found)
		require.Equal(t, item, got)
	}
}

func TestHostZoneRemove(t *testing.T) {
	app := app.Setup(t)
	ctx := app.NewContextLegacy(true, cmtproto.Header{Height: 1})

	expected := createNHostZone(t, &app.FeeabsKeeper, ctx, 1)
	for _, item := range expected {
		err := app.FeeabsKeeper.DeleteHostZoneConfig(ctx, item.IbcDenom)
		require.NoError(t, err)
		_, found := app.FeeabsKeeper.GetHostZoneConfig(ctx, item.IbcDenom)
		require.False(t, found)
		_, found = app.FeeabsKeeper.GetHostZoneConfigByOsmosisTokenDenom(ctx, item.OsmosisPoolTokenDenomIn)
		require.False(t, found)
	}
}

func TestHostZoneGetAll(t *testing.T) {
	app := app.Setup(t)
	ctx := app.NewContextLegacy(true, cmtproto.Header{Height: 1})

	expected := createNHostZone(t, &app.FeeabsKeeper, ctx, 1)
	got, _ := app.FeeabsKeeper.GetAllHostZoneConfig(ctx)
	require.ElementsMatch(t, expected, got)
}
