package interchaintest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	paramsutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"

	feeabsCli "github.com/osmosis-labs/fee-abstraction/v7/tests/interchaintest/feeabs"
)

func TestQueryOsmosisTwap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	// Set up chains, users and channels
	ctx := context.Background()
	chains, users, channels := SetupChain(t, ctx)
	feeabs, gaia, osmosis := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)

	feeabsUser, _, osmosisUser := users[0], users[1], users[2]

	channFeeabsOsmosis, channOsmosisFeeabs, channFeeabsGaia, channGaiaFeeabs, channOsmosisGaia, channGaiaOsmosis, channFeeabsOsmosisICQ, _ := channels[0], channels[1], channels[2], channels[3], channels[4], channels[5], channels[6], channels[7]

	// Setup contract on Osmosis
	// Store code crosschain Registry
	crossChainRegistryContractID, err := osmosis.StoreContract(ctx, osmosisUser.KeyName(), "./bytecode/crosschain_registry.wasm")
	require.NoError(t, err)
	_ = crossChainRegistryContractID
	// // Instatiate
	owner := sdktypes.MustBech32ifyAddressBytes(osmosis.Config().Bech32Prefix, osmosisUser.Address())
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
	osmosisUserBalance, err := osmosis.GetBalance(ctx, sdktypes.MustBech32ifyAddressBytes(osmosis.Config().Bech32Prefix, osmosisUser.Address()), stakeOnOsmosis)
	require.NoError(t, err)
	require.Equal(t, amountToSend, osmosisUserBalance)

	poolID, err := feeabsCli.CreatePool(osmosis, ctx, osmosisUser.KeyName(), cosmos.OsmosisPoolParams{
		Weights:        fmt.Sprintf("5%s,5%s", stakeOnOsmosis, osmosis.Config().Denom),
		InitialDeposit: fmt.Sprintf("95000000%s,950000000%s", stakeOnOsmosis, osmosis.Config().Denom),
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

	ParamChangeProposal(t, ctx, feeabs, feeabsUser, &channFeeabsOsmosis, &channFeeabsOsmosisICQ, stakeOnOsmosis)
	AddHostZoneProposal(t, ctx, feeabs, feeabsUser)
	// ensure that the host zone is added
	allHost, err := feeabsCli.QueryAllHostZoneConfig(feeabs, ctx)
	require.NoError(t, err)
	fmt.Printf("QueryAllHostZoneConfig %+v", allHost)

	// try to query both via osmosis client and by interchainquery
	err = testutil.WaitForBlocks(ctx, 15, feeabs)
	require.NoError(t, err)

	twapOsmosis, err := feeabsCli.QueryOsmosisArithmeticTwap(feeabs, ctx, stakeOnOsmosis)
	require.NoError(t, err)
	fmt.Println(twapOsmosis)

	twap, err := feeabsCli.QueryOsmosisArithmeticTwapOsmosis(osmosis, ctx, "1", stakeOnOsmosis)
	fmt.Println(twap)
	require.NoError(t, err)
}

func ParamChangeProposal(t *testing.T, ctx context.Context, feeabs *cosmos.CosmosChain, feeabsUser ibc.Wallet, channFeeabsOsmosis, channFeeabsOsmosisFeeabs *ibc.ChannelOutput, stakeOnOsmosis string) {
	t.Helper()
	// propose to change feeabs parameters accordingly to the ibcdenom
	curDir, _ := os.Getwd()
	paramChangePath := path.Join(curDir, "proposal", "proposal.json")

	changeParamProposal, err := paramsutils.ParseParamChangeProposalJSON(feeabs.Config().EncodingConfig.Amino, paramChangePath)
	require.NoError(t, err)

	// modify change proposal
	for i := range changeParamProposal.Changes {
		change := &changeParamProposal.Changes[i]
		if change.Subspace == "feeabs" && change.Key == "IbcTransferChannel" {
			fmt.Println("ibc transfer channel changed", channFeeabsOsmosis.ChannelID)
			change.Value = json.RawMessage(fmt.Sprintf("\"%s\"", channFeeabsOsmosis.ChannelID))
		}
		if change.Subspace == "feeabs" && change.Key == "IbcQueryIcqChannel" {
			fmt.Println("ibc query icq channel changed", channFeeabsOsmosisFeeabs.ChannelID)
			change.Value = json.RawMessage(fmt.Sprintf("\"%s\"", channFeeabsOsmosisFeeabs.ChannelID))
		}
		if change.Subspace == "feeabs" && change.Key == "NativeIbcedInOsmosis" {
			fmt.Println("NativeIbcedInOsmosis changed", stakeOnOsmosis)
			change.Value = json.RawMessage(fmt.Sprintf("\"%s\"", stakeOnOsmosis))
		}
	}
	fmt.Printf("changeParamProposal %+v", changeParamProposal)

	paramTx, err := feeabsCli.ParamChangeProposal(feeabs, ctx, feeabsUser.KeyName(), &changeParamProposal)
	require.NoError(t, err, "error submitting param change proposal tx")

	err = feeabs.VoteOnProposalAllValidators(ctx, paramTx.ProposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	height, err := feeabs.Height(ctx)
	require.NoError(t, err)

	_, err = cosmos.PollForProposalStatus(ctx, feeabs, height, height+20, paramTx.ProposalID, cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")
}

func AddHostZoneProposal(t *testing.T, ctx context.Context, feeabs *cosmos.CosmosChain, feeabsUser ibc.Wallet) {
	t.Helper()
	_, err := feeabsCli.AddHostZoneProposal(feeabs, ctx, feeabsUser.KeyName(), "./proposal/add_host_zone.json")
	require.NoError(t, err)

	err = feeabs.VoteOnProposalAllValidators(ctx, "2", cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	height, err := feeabs.Height(ctx)
	require.NoError(t, err)

	_, err = cosmos.PollForProposalStatus(ctx, feeabs, height, height+20, "2", cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")
}

func GetStakeOnOsmosis(channOsmosisFeeabs ibc.ChannelOutput, feeabsDenom string) string {
	denomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(channOsmosisFeeabs.PortID, channOsmosisFeeabs.ChannelID, feeabsDenom))
	stakeOnOsmosis := denomTrace.IBCDenom()
	return stakeOnOsmosis
}
