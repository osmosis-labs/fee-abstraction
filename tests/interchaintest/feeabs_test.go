package interchaintest

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	paramsutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"

	feeabsCli "github.com/osmosis-labs/fee-abstraction/tests/interchaintest/feeabs"
)

func TestFeeAbs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	// Set up chains, users and channels
	ctx := context.Background()
	chains, users, channels := SetupChain(t, ctx)
	feeabs, gaia, osmosis := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)

	feeabsUser, gaiaUser, osmosisUser := users[0], users[1], users[2]

	channFeeabsOsmosis, channOsmosisFeeabs, channFeeabsGaia, channGaiaFeeabs, channOsmosisGaia, channGaiaOsmosis := channels[0], channels[1], channels[2], channels[3], channels[4], channels[5]

	contracts, err := SetupOsmosisContracts(t, ctx, osmosis, osmosisUser)
	require.NoError(t, err)
	require.Equal(t, len(contracts), 3)

	registryContractAddr := contracts[0]
	swapRouterContractAddr := contracts[1]
	_ = contracts[2]

	// Modify chain channel links on registry contract
	msg := fmt.Sprintf("{\"modify_chain_channel_links\": {\"operations\": [{\"operation\": \"set\",\"source_chain\": \"feeabs\",\"destination_chain\": \"osmosis\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"osmosis\",\"destination_chain\": \"feeabs\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"feeabs\",\"destination_chain\": \"gaia\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"gaia\",\"destination_chain\": \"feeabs\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"osmosis\",\"destination_chain\": \"gaia\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"gaia\",\"destination_chain\": \"osmosis\",\"channel_id\": \"%s\"}]}}",
		channFeeabsOsmosis.ChannelID,
		channOsmosisFeeabs.ChannelID,
		channFeeabsGaia.ChannelID,
		channGaiaFeeabs.ChannelID,
		channOsmosisGaia.ChannelID,
		channGaiaOsmosis.ChannelID)
	_, err = osmosis.ExecuteContract(ctx, osmosisUser.KeyName(), registryContractAddr, msg, "--gas", "1000000")
	require.NoError(t, err)

	// Modify bech32 prefixes on registry contract
	msg = `{
			"modify_bech32_prefixes": 
			{
				"operations": 
				[
					{"operation": "set", "chain_name": "feeabs", "prefix": "feeabs"},
					{"operation": "set", "chain_name": "osmosis", "prefix": "osmo"},
					{"operation": "set", "chain_name": "gaia", "prefix": "cosmos"}
				]
			}
		}`
	_, err = osmosis.ExecuteContract(ctx, osmosisUser.KeyName(), registryContractAddr, msg)
	require.NoError(t, err)

	osmosisPrefix := osmosis.Config().Bech32Prefix

	uatomOnOsmosis := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(channOsmosisGaia.PortID, channOsmosisGaia.ChannelID, gaia.Config().Denom)).IBCDenom()
	osmosisUserBalance, err := osmosis.GetBalance(
		ctx,
		sdktypes.MustBech32ifyAddressBytes(osmosisPrefix, osmosisUser.Address()),
		uatomOnOsmosis,
	)
	require.NoError(t, err)
	require.Equal(t, amountToSend, osmosisUserBalance)

	stakeOnOsmosis := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(channOsmosisFeeabs.PortID, channOsmosisFeeabs.ChannelID, feeabs.Config().Denom)).IBCDenom()
	osmosisUserBalance, err = osmosis.GetBalance(
		ctx,
		sdktypes.MustBech32ifyAddressBytes(osmosisPrefix, osmosisUser.Address()),
		stakeOnOsmosis,
	)
	require.NoError(t, err)
	require.Equal(t, amountToSend, osmosisUserBalance)

	// Setup propose_pfm
	// propose_pfm for feeabs
	_, err = feeabsCli.SetupProposePFM(osmosis, ctx, osmosisUser.KeyName(), registryContractAddr, `{"propose_pfm":{"chain": "feeabs"}}`, stakeOnOsmosis)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 15, feeabs, gaia, osmosis)
	require.NoError(t, err)

	queryMsg := QuerySmartMsg{
		Packet: HasPacketForwarding{
			Chain: "feeabs",
		},
	}
	// {"data":false}
	var feeabsRes QuerySmartMsgResponse
	err = osmosis.QueryContract(ctx, registryContractAddr, queryMsg, &feeabsRes)
	require.NoError(t, err)
	require.Equal(t, true, feeabsRes.Data)

	// propose_pfm for gaia
	_, err = feeabsCli.SetupProposePFM(osmosis, ctx, osmosisUser.KeyName(), registryContractAddr, `{"propose_pfm":{"chain": "gaia"}}`, uatomOnOsmosis)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 15, feeabs, gaia, osmosis)
	require.NoError(t, err)

	queryMsg = QuerySmartMsg{
		Packet: HasPacketForwarding{
			Chain: "gaia",
		},
	}
	var gaiaRes QuerySmartMsgResponse
	err = osmosis.QueryContract(ctx, registryContractAddr, queryMsg, &gaiaRes)
	require.NoError(t, err)
	require.Equal(t, true, gaiaRes.Data)

	// Create pool uatom/stake on Osmosis
	poolID, err := feeabsCli.CreatePool(osmosis, ctx, osmosisUser.KeyName(), cosmos.OsmosisPoolParams{
		Weights:        fmt.Sprintf("5%s,5%s", stakeOnOsmosis, uatomOnOsmosis),
		InitialDeposit: fmt.Sprintf("95000000%s,950000000%s", stakeOnOsmosis, uatomOnOsmosis),
		SwapFee:        "0.01",
		ExitFee:        "0",
		FutureGovernor: "",
	})
	require.NoError(t, err)
	require.Equal(t, "1", poolID)

	// execute
	msg = fmt.Sprintf("{\"set_route\":{\"input_denom\":\"%s\",\"output_denom\":\"%s\",\"pool_route\":[{\"pool_id\":\"%s\",\"token_out_denom\":\"%s\"}]}}",
		uatomOnOsmosis,
		stakeOnOsmosis,
		poolID,
		stakeOnOsmosis,
	)
	_, err = osmosis.ExecuteContract(ctx, osmosisUser.KeyName(), swapRouterContractAddr, msg)
	require.NoError(t, err)

	// Swap Feeabs(uatom) to Osmosis
	// send ibc token to feeabs module account
	gaiaHeight, err := gaia.Height(ctx)
	require.NoError(t, err)

	feeabsModule, err := feeabsCli.QueryModuleAccountBalances(feeabs, ctx)
	require.NoError(t, err)
	transfer := ibc.WalletAmount{
		Address: feeabsModule.GetAddress(),
		Denom:   gaia.Config().Denom,
		Amount:  math.NewInt(1000000),
	}

	ibcTx, err := gaia.SendIBCTransfer(ctx, channGaiaFeeabs.ChannelID, gaiaUser.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)
	require.NoError(t, ibcTx.Validate())

	_, err = testutil.PollForAck(ctx, gaia, gaiaHeight, gaiaHeight+30, ibcTx.Packet)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 1, feeabs, gaia, osmosis)
	require.NoError(t, err)

	uatomOnFeeabs := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(channFeeabsGaia.PortID, channFeeabsGaia.ChannelID, gaia.Config().Denom)).IBCDenom()

	// Proposal change params of feeabs
	currentDirectory, _ := os.Getwd()
	paramChangePath := path.Join(currentDirectory, "proposal", "proposal.json")

	changeParamProposal, err := paramsutils.ParseParamChangeProposalJSON(feeabs.Config().EncodingConfig.Amino, paramChangePath)
	require.NoError(t, err)

	paramTx, err := feeabsCli.ParamChangeProposal(feeabs, ctx, feeabsUser.KeyName(), &changeParamProposal)
	require.NoError(t, err, "error submitting param change proposal tx")

	err = feeabs.VoteOnProposalAllValidators(ctx, paramTx.ProposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	height, err := feeabs.Height(ctx)
	require.NoError(t, err)

	_, err = cosmos.PollForProposalStatus(ctx, feeabs, height, height+10, paramTx.ProposalID, cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	_, err = feeabsCli.AddHostZoneProposal(feeabs, ctx, feeabsUser.KeyName(), "./proposal/add_host_zone.json")
	require.NoError(t, err)

	err = feeabs.VoteOnProposalAllValidators(ctx, "2", cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	height, err = feeabs.Height(ctx)
	require.NoError(t, err)

	_, err = cosmos.PollForProposalStatus(ctx, feeabs, height, height+10, paramTx.ProposalID, cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	// wait for next 5 blocks
	require.NoError(t, err)
	testutil.WaitForBlocks(ctx, 5, feeabs)

	// there must be exactly 1 host zone configs
	res, err := feeabsCli.QueryAllHostZoneConfig(feeabs, ctx)
	require.NoError(t, err)
	require.Equal(t, len(res.AllHostChainConfig), 1)

	// xcs
	feeabsHeight, err := feeabs.Height(ctx)
	require.NoError(t, err)

	feeabsModule, err = feeabsCli.QueryModuleAccountBalances(feeabs, ctx)
	require.NoError(t, err)
	t.Logf("Module Account Balances before swap: %v\n", feeabsModule.Balances)

	transferTx, err := feeabsCli.CrossChainSwap(feeabs, ctx, feeabsUser.KeyName(), uatomOnFeeabs)
	require.NoError(t, err)
	_, err = testutil.PollForAck(ctx, feeabs, feeabsHeight, feeabsHeight+25, transferTx.Packet)
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, 50, feeabs, gaia, osmosis)
	require.NoError(t, err)

	feeabsModule, err = feeabsCli.QueryModuleAccountBalances(feeabs, ctx)
	require.NoError(t, err)
	t.Logf("Module Account Balances after swap: %v\n", feeabsModule.Balances)

	balance, err := feeabs.GetBalance(ctx, feeabsModule.Address, feeabs.Config().Denom)
	require.NoError(t, err)
	require.True(t, balance.GT(math.NewInt(1)))
}
