package interchaintest

import (
	"context"
	"fmt"
	"testing"

	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"

	feeabsCli "github.com/osmosis-labs/fee-abstraction/tests/interchaintest/feeabs"
)

func TestHostZoneProposal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	ctx := context.Background()

	chains, users, _ := SetupChain(t, ctx)
	feeabs, _, _ := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)

	feeabsUser, _, _ := users[0], users[1], users[2]

	// Start testing for add host zone proposal
	_, err := feeabsCli.AddHostZoneProposal(feeabs, ctx, feeabsUser.KeyName(), "./proposal/add_host_zone.json")
	require.NoError(t, err)

	err = feeabs.VoteOnProposalAllValidators(ctx, "1", cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	height, err := feeabs.Height(ctx)
	require.NoError(t, err)

	_, err = cosmos.PollForProposalStatus(ctx, feeabs, height, height+10, 1, govv1beta1.StatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	config, err := feeabsCli.QueryHostZoneConfigWithDenom(feeabs, ctx, "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9")
	require.NoError(t, err)
	require.Equal(t, config, &feeabsCli.HostChainFeeAbsConfigResponse{HostChainConfig: feeabsCli.HostChainFeeAbsConfig{
		IbcDenom:                "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9",
		OsmosisPoolTokenDenomIn: "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9",
		PoolId:                  "1",
		Status:                  feeabsCli.HostChainFeeAbsStatus_UPDATED,
	}})

	// Start testing for set host zone proposal
	_, err = feeabsCli.SetHostZoneProposal(feeabs, ctx, feeabsUser.KeyName(), "./proposal/set_host_zone.json")
	require.NoError(t, err)

	err = feeabs.VoteOnProposalAllValidators(ctx, "2", cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	height, err = feeabs.Height(ctx)
	require.NoError(t, err)

	_, err = cosmos.PollForProposalStatus(ctx, feeabs, height, height+10, 2, govv1beta1.StatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	config, err = feeabsCli.QueryHostZoneConfigWithDenom(feeabs, ctx, "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9")
	require.NoError(t, err)
	require.Equal(t, config, &feeabsCli.HostChainFeeAbsConfigResponse{HostChainConfig: feeabsCli.HostChainFeeAbsConfig{
		IbcDenom:                "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9",
		OsmosisPoolTokenDenomIn: "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9",
		PoolId:                  "1",
		Status:                  feeabsCli.HostChainFeeAbsStatus_FROZEN,
	}})

	// Start testing for delete host zone proposal
	_, err = feeabsCli.DeleteHostZoneProposal(feeabs, ctx, feeabsUser.KeyName(), "./proposal/delete_host_zone.json")
	require.NoError(t, err)

	err = feeabs.VoteOnProposalAllValidators(ctx, "3", cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	height, err = feeabs.Height(ctx)
	require.NoError(t, err)

	response, err := cosmos.PollForProposalStatus(ctx, feeabs, height, height+10, 3, govv1beta1.StatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")
	fmt.Printf("response: %s\n", response)

	_, err = feeabsCli.QueryHostZoneConfigWithDenom(feeabs, ctx, "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9")
	require.Error(t, err) // not found
}
