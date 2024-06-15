package interchaintest

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	feeabsCli "github.com/osmosis-labs/fee-abstraction/v8/tests/interchaintest/feeabs"
	feeabstypes "github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

func TestQueryOsmosisTwap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	// Set up chains, users and channels
	ctx := context.Background()
	chains, users, channels := SetupChain(t, ctx)
	feeabs, gaia, osmosis := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)

	feeabsUser, _, osmosisUser := users[0], users[1], users[2]

	channFeeabsOsmosis, channOsmosisFeeabs, channFeeabsGaia, channGaiaFeeabs, channOsmosisGaia, channGaiaOsmosis, channFeeabsOsmosisICQ := channels[0], channels[1], channels[2], channels[3], channels[4], channels[5], channels[6]

	// Setup contract on Osmosis
	// Store code crosschain Registry
	crossChainRegistryContractID, err := osmosis.StoreContract(ctx, osmosisUser.KeyName(), "./bytecode/crosschain_registry.wasm")
	require.NoError(t, err)
	_ = crossChainRegistryContractID
	// // Instatiate
	owner := sdk.MustBech32ifyAddressBytes(osmosis.Config().Bech32Prefix, osmosisUser.Address())
	initMsg := fmt.Sprintf("{\"owner\":\"%s\"}", owner)
	registryContractAddress, err := osmosis.InstantiateContract(ctx, osmosisUser.KeyName(), crossChainRegistryContractID, initMsg, true)
	require.NoError(t, err)
	// Execute
	msg := fmt.Sprintf("{\"modify_chain_channel_links\": {\"operations\": [{\"operation\": \"set\",\"source_chain\": \"feeabs\",\"destination_chain\": \"osmosis\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"osmosis\",\"destination_chain\": \"feeabs\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"feeabs\",\"destination_chain\": \"gaia\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"gaia\",\"destination_chain\": \"feeabs\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"osmosis\",\"destination_chain\": \"gaia\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"gaia\",\"destination_chain\": \"osmosis\",\"channel_id\": \"%s\"}]}}",
		channFeeabsOsmosis.ChannelID,
		channOsmosisFeeabs.ChannelID,
		channFeeabsGaia.ChannelID,
		channGaiaFeeabs.ChannelID,
		channOsmosisGaia.ChannelID,
		channGaiaOsmosis.ChannelID)
	_, err = osmosis.ExecuteContract(ctx, osmosisUser.KeyName(), registryContractAddress, msg, "--gas", "1000000")
	require.NoError(t, err)
	// Execute
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
	_, err = osmosis.ExecuteContract(ctx, osmosisUser.KeyName(), registryContractAddress, msg)
	require.NoError(t, err)

	// Create pool Osmosis(stake)/uosmo on Osmosis
	stakeOnOsmosis := GetStakeOnOsmosis(channOsmosisFeeabs, feeabs.Config().Denom)
	osmosisUserBalance, err := osmosis.GetBalance(ctx, sdk.MustBech32ifyAddressBytes(osmosis.Config().Bech32Prefix, osmosisUser.Address()), stakeOnOsmosis)
	require.NoError(t, err)
	require.Equal(t, amountToSend, osmosisUserBalance)

	initAmount := amountToSend.Uint64() / 2
	poolID, err := feeabsCli.CreatePool(osmosis, ctx, osmosisUser.KeyName(), cosmos.OsmosisPoolParams{
		Weights:        fmt.Sprintf("5%s,5%s", stakeOnOsmosis, osmosis.Config().Denom),
		InitialDeposit: fmt.Sprintf("%d%s,%d%s", initAmount, stakeOnOsmosis, initAmount, osmosis.Config().Denom),
		SwapFee:        "0.01",
		ExitFee:        "0",
		FutureGovernor: "",
	})
	require.NoError(t, err)
	require.Equal(t, poolID, "1")

	// Setup propose_pfm
	// propose_pfm for feeabs
	_, err = feeabsCli.SetupProposePFM(osmosis, ctx, osmosisUser.KeyName(), registryContractAddress, `{"propose_pfm":{"chain": "feeabs"}}`, stakeOnOsmosis)
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, 15, feeabs, gaia, osmosis)
	require.NoError(t, err)
	queryMsg := QuerySmartMsg{
		Packet: HasPacketForwarding{
			Chain: "feeabs",
		},
	}
	res := QuerySmartMsgResponse{}
	err = osmosis.QueryContract(ctx, registryContractAddress, queryMsg, &res)
	require.NoError(t, err)

	ParamChangeProposal(t, ctx, feeabs, feeabsUser, channFeeabsOsmosis.ChannelID, channFeeabsOsmosisICQ.ChannelID, stakeOnOsmosis)
	AddHostZoneProposal(t, ctx, feeabs, feeabsUser)

	// ensure that the host zone is added
	allHost, err := feeabsCli.QueryAllHostZoneConfig(feeabs, ctx)
	require.NoError(t, err)
	fmt.Printf("QueryAllHostZoneConfig %+v", allHost)
	err = testutil.WaitForBlocks(ctx, 15, feeabs)
	require.NoError(t, err)

	// query the twap of uosmo/stake, stored in feeabs module
	osmoOnFeeabs := GetOsmoOnFeeabs(channFeeabsOsmosis, osmosis.Config().Denom)
	twapOsmosis, err := feeabsCli.QueryOsmosisArithmeticTwap(feeabs, ctx, osmoOnFeeabs)
	require.NoError(t, err)
	fmt.Println(twapOsmosis)

	// query the twap of uosmo/stake
	twap, err := feeabsCli.QueryOsmosisArithmeticTwapOsmosis(osmosis, ctx, "1", stakeOnOsmosis)
	fmt.Println(twap)
	require.NoError(t, err)
}

func ParamChangeProposal(
	t *testing.T,
	ctx context.Context,
	feeabs *cosmos.CosmosChain,
	feeabsUser ibc.Wallet,
	channFeeabsOsmosis, channFeeabsOsmosisFeeabs string,
	stakeOnOsmosis string,
) {
	t.Helper()
	govAddr, err := feeabs.AuthQueryModuleAddress(ctx, "gov")
	require.NoError(t, err)
	require.NotEmpty(t, govAddr)
	updateParamMsg := feeabstypes.MsgUpdateParams{
		Params: feeabstypes.Params{
			OsmosisQueryTwapPath:         "/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow",
			IbcTransferChannel:           channFeeabsOsmosis,
			IbcQueryIcqChannel:           channFeeabsOsmosisFeeabs,
			NativeIbcedInOsmosis:         stakeOnOsmosis,
			OsmosisCrosschainSwapAddress: "osmo17p9rzwnnfxcjp32un9ug7yhhzgtkhvl9jfksztgw5uh69wac2pgs5yczr8",
			ChainName:                    feeabs.Config().ChainID,
		},
		Authority: govAddr,
	}

	title := "Test Proposal"
	prop, err := feeabs.BuildProposal([]cosmos.ProtoMessage{&updateParamMsg}, title, title+" Summary", "none", "5000000000"+feeabs.Config().Denom, govAddr, false)
	fmt.Printf("prop %+v", prop)

	require.NoError(t, err, "error building param change proposal")
	paramTx, err := feeabs.SubmitProposal(ctx, feeabsUser.KeyName(), prop)
	require.NoError(t, err, "error submitting param change proposal tx")

	proposalID, err := strconv.ParseUint(paramTx.ProposalID, 10, 64)
	require.NoError(t, err, "parse proposal id failed")

	err = feeabs.VoteOnProposalAllValidators(ctx, proposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	height, err := feeabs.Height(ctx)
	require.NoError(t, err)

	propID, err := strconv.ParseUint(paramTx.ProposalID, 10, 64)
	require.NoError(t, err, "parse proposal id failed")

	_, err = cosmos.PollForProposalStatus(ctx, feeabs, height, height+20, propID, v1beta1.StatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")
}

func AddHostZoneProposal(t *testing.T, ctx context.Context, feeabs *cosmos.CosmosChain, feeabsUser ibc.Wallet) {
	t.Helper()

	govAddr, err := feeabs.AuthQueryModuleAddress(ctx, "gov")
	require.NoError(t, err)
	require.NotEmpty(t, govAddr)

	addHostZoneMsg := feeabstypes.MsgAddHostZone{
		HostChainConfig: feeabstypes.HostChainFeeAbsConfig{
			IbcDenom:                fakeIBCDenom,
			OsmosisPoolTokenDenomIn: "uosmo",
			PoolId:                  1,
			Status:                  0,
		},
		Authority: govAddr,
	}
	title := "Test Proposal"
	prop, err := feeabs.BuildProposal([]cosmos.ProtoMessage{&addHostZoneMsg}, title, title+" Summary", "none", "5000000000"+feeabs.Config().Denom, govAddr, false)
	fmt.Printf("prop %+v", prop)
	require.NoError(t, err)

	proposalTx, err := feeabs.SubmitProposal(ctx, feeabsUser.KeyName(), prop)
	require.NoError(t, err, "error submitting param change proposal tx")

	proposalID, err := strconv.ParseUint(proposalTx.ProposalID, 10, 64)
	require.NoError(t, err, "parse proposal id failed")

	err = feeabs.VoteOnProposalAllValidators(ctx, proposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	height, err := feeabs.Height(ctx)
	require.NoError(t, err)

	_, err = cosmos.PollForProposalStatus(ctx, feeabs, height, height+20, proposalID, v1beta1.StatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")
}

func GetStakeOnOsmosis(channOsmosisFeeabs ibc.ChannelOutput, feeabsDenom string) string {
	denomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(channOsmosisFeeabs.PortID, channOsmosisFeeabs.ChannelID, feeabsDenom))
	stakeOnOsmosis := denomTrace.IBCDenom()
	return stakeOnOsmosis
}

func GetOsmoOnFeeabs(channFeeabsOsmosis ibc.ChannelOutput, osmosisDenom string) string {
	denomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(channFeeabsOsmosis.PortID, channFeeabsOsmosis.ChannelID, osmosisDenom))
	osmoOnFeeabs := denomTrace.IBCDenom()
	return osmoOnFeeabs
}
