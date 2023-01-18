package keeper_test

import (
	"testing"

	apphelpers "github.com/notional-labs/feeabstraction/v1/app/helpers"
	feeabstypes "github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
	"github.com/stretchr/testify/require"
)

func TestGetOsmosisIBCDenomParams(t *testing.T) {
	app := apphelpers.Setup(t, false, 1)
	ctx := apphelpers.NewContextForApp(*app)

	params := feeabstypes.Params{
		OsmosisIbcDenom: "ibc/acb",
	}
	app.FeeabsKeeper.SetParams(ctx, params)

	osmosisIBCDenom := app.FeeabsKeeper.GetOsmosisIBCDenomParams(ctx)
	require.Equal(t, params.OsmosisIbcDenom, osmosisIBCDenom)
}
