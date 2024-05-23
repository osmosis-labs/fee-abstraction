package interchaintest

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/stretchr/testify/require"

	feeabsCli "github.com/osmosis-labs/fee-abstraction/v7/tests/interchaintest/feeabs"
)

func TestHostZoneProposal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	ctx := context.Background()

	chains, users, channels := SetupChain(t, ctx)
	feeabs, _, _ := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)
	channFeeabsOsmosis, channOsmosisFeeabs, channFeeabsOsmosisICQ := channels[0], channels[1], channels[6]

	feeabsUser, _, _ := users[0], users[1], users[2]
	stakeOnOsmosis := GetStakeOnOsmosis(channOsmosisFeeabs, feeabs.Config().Denom)

	ParamChangeProposal(t, ctx, feeabs, feeabsUser, &channFeeabsOsmosis, &channFeeabsOsmosisICQ, stakeOnOsmosis)
	AddHostZoneProposal(t, ctx, feeabs, feeabsUser)

	_, err := feeabsCli.QueryHostZoneConfigWithDenom(feeabs, ctx, stakeOnOsmosis)
	require.NoError(t, err)
}
