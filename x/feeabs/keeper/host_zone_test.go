package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	apphelpers "github.com/osmosis-labs/fee-abstraction/v7/app/helpers"
	feeabskeeper "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/keeper"
	"github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
)

func createNHostZone(t *testing.T, keeper *feeabskeeper.Keeper, ctx sdk.Context, n int) []types.HostChainFeeAbsConfig {
	t.Helper()
	var expected []types.HostChainFeeAbsConfig
	expectedConfig := types.HostChainFeeAbsConfig{
		IbcDenom:                "ibc/123",
		OsmosisPoolTokenDenomIn: "ibc/456",
		PoolRoute: []types.PoolRoute{
			{
				PoolId:        1,
				TokenOutDenom: "uosmo",
			},
		},
		Frozen: false,
	}
	for i := 0; i < n; i++ {
		expected = append(expected, expectedConfig)
		err := keeper.SetHostZoneConfig(ctx, expectedConfig.IbcDenom, expectedConfig)
		require.NoError(t, err)
	}
	return expected
}

func TestHostZoneGet(t *testing.T) {
	app := apphelpers.Setup(t, false, 1)
	ctx := apphelpers.NewContextForApp(*app)
	expected := createNHostZone(t, &app.FeeabsKeeper, ctx, 1)
	for _, item := range expected {
		got, found := app.FeeabsKeeper.GetHostZoneConfig(ctx, item.IbcDenom)
		require.True(t, found)
		require.Equal(t, item, got)
	}
}

func TestHostZoneRemove(t *testing.T) {
	app := apphelpers.Setup(t, false, 1)
	ctx := apphelpers.NewContextForApp(*app)
	expected := createNHostZone(t, &app.FeeabsKeeper, ctx, 1)
	for _, item := range expected {
		err := app.FeeabsKeeper.DeleteHostZoneConfig(ctx, item.IbcDenom)
		require.NoError(t, err)
		got, _ := app.FeeabsKeeper.GetHostZoneConfig(ctx, item.IbcDenom)
		require.NotEqual(t, item, got)
	}
}

func TestHostZoneGetAll(t *testing.T) {
	app := apphelpers.Setup(t, false, 1)
	ctx := apphelpers.NewContextForApp(*app)
	expected := createNHostZone(t, &app.FeeabsKeeper, ctx, 1)
	got, _ := app.FeeabsKeeper.GetAllHostZoneConfig(ctx)
	require.ElementsMatch(t, expected, got)
}
