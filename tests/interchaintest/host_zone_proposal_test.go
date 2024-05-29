package interchaintest

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"

	feeabsCli "github.com/osmosis-labs/fee-abstraction/v8/tests/interchaintest/feeabs"
)

func TestHostZoneProposal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	ctx := context.Background()

	chains, users, channels := SetupChain(t, ctx)
	feeabs, _, osmosis := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)
	channFeeabsOsmosis, _, channFeeabsOsmosisICQ := channels[0], channels[1], channels[6]

	feeabsUser, _, _ := users[0], users[1], users[2]
	osmoOnFeeabs := GetOsmoOnFeeabs(channFeeabsOsmosis, osmosis.Config().Denom)

	ParamChangeProposal(t, ctx, feeabs, feeabsUser, &channFeeabsOsmosis, &channFeeabsOsmosisICQ, osmoOnFeeabs)
	AddHostZoneProposal(t, ctx, feeabs, feeabsUser)

	_, err := feeabsCli.QueryHostZoneConfigWithDenom(feeabs, ctx, osmoOnFeeabs)
	require.NoError(t, err)
}
