package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	apphelpers "github.com/osmosis-labs/fee-abstraction/v8/app/helpers"
	feeabskeeper "github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/keeper"
	"github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

func createEpoch(t *testing.T, keeper *feeabskeeper.Keeper, ctx sdk.Context) types.EpochInfo {
	t.Helper()
	expected := types.EpochInfo{
		Identifier:              "Test",
		StartTime:               time.Now().UTC(),
		Duration:                10,
		CurrentEpoch:            0,
		CurrentEpochStartTime:   time.Now().UTC(),
		EpochCountingStarted:    false,
		CurrentEpochStartHeight: 0,
	}
	err := keeper.AddEpochInfo(ctx, expected)
	require.NoError(t, err)

	err = keeper.AddEpochInfo(ctx, expected)
	require.Error(t, err, "epoch with identifier Test already exists")

	return expected
}

func TestGetEpochInfo(t *testing.T) {
	app := apphelpers.Setup(t, false, 1)
	ctx := apphelpers.NewContextForApp(*app)
	expected := createEpoch(t, &app.FeeabsKeeper, ctx)
	got, found := app.FeeabsKeeper.GetEpochInfo(ctx, expected.Identifier)
	require.True(t, found)
	require.Equal(t, expected.StartTime, got.StartTime)
	require.Equal(t, expected.Duration, got.Duration)
	require.Equal(t, expected.EpochCountingStarted, got.EpochCountingStarted)
}

func TestHasEpochInfo(t *testing.T) {
	app := apphelpers.Setup(t, false, 1)
	ctx := apphelpers.NewContextForApp(*app)
	expected := createEpoch(t, &app.FeeabsKeeper, ctx)
	found := app.FeeabsKeeper.HasEpochInfo(ctx, expected.Identifier)
	require.True(t, found)
}
