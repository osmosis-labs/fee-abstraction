package interchaintest

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	feeabsCli "github.com/osmosis-labs/fee-abstraction/v7/tests/interchaintest/feeabs"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// TestFeeabsGaiaIBCTransfer spins up a Feeabs and Gaia network, initializes an IBC connection between them,
// and sends an ICS20 token transfer from Feeabs->Gaia and then back from Gaia->Feeabs.
func TestFeeabsGaiaIBCTransferWithIBCFee(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	// Set up chains, users and channels
	ctx := context.Background()
	chains, users, channels := SetupChain(t, ctx)
	feeabs, gaia, osmosis := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)

	feeabsUser, _, osmosisUser := users[0], users[1], users[2]

	channFeeabsOsmosis, channOsmosisFeeabs, channFeeabsGaia, channGaiaFeeabs, channOsmosisGaia, channGaiaOsmosis := channels[0], channels[1], channels[2], channels[3], channels[4], channels[5]

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

	// Create pool Osmosis(uatom)/Osmosis(stake) on Osmosis
	denomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(channOsmosisGaia.PortID, channOsmosisGaia.ChannelID, gaia.Config().Denom))
	uatomOnOsmosis := denomTrace.IBCDenom()
	osmosisUserBalance, err := osmosis.GetBalance(ctx, sdktypes.MustBech32ifyAddressBytes(osmosis.Config().Bech32Prefix, osmosisUser.Address()), uatomOnOsmosis)
	require.NoError(t, err)
	require.Equal(t, amountToSend, osmosisUserBalance)

	denomTrace = transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(channOsmosisFeeabs.PortID, channOsmosisFeeabs.ChannelID, feeabs.Config().Denom))
	stakeOnOsmosis := denomTrace.IBCDenom()
	osmosisUserBalance, err = osmosis.GetBalance(ctx, sdktypes.MustBech32ifyAddressBytes(osmosis.Config().Bech32Prefix, osmosisUser.Address()), stakeOnOsmosis)
	require.NoError(t, err)
	require.Equal(t, amountToSend, osmosisUserBalance)

	poolID, err := feeabsCli.CreatePool(osmosis, ctx, osmosisUser.KeyName(), cosmos.OsmosisPoolParams{
		Weights:        fmt.Sprintf("5%s,5%s", osmosis.Config().Denom, stakeOnOsmosis),
		InitialDeposit: fmt.Sprintf("95000000%s,950000000%s", osmosis.Config().Denom, stakeOnOsmosis),
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
	// propose_pfm for gaia
	_, err = feeabsCli.SetupProposePFM(osmosis, ctx, osmosisUser.KeyName(), registryContractAddress, `{"propose_pfm":{"chain": "gaia"}}`, uatomOnOsmosis)
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, 15, feeabs, gaia, osmosis)
	require.NoError(t, err)
	queryMsg = QuerySmartMsg{
		Packet: HasPacketForwarding{
			Chain: "gaia",
		},
	}
	res = QuerySmartMsgResponse{}
	err = osmosis.QueryContract(ctx, registryContractAddress, queryMsg, &res)
	require.NoError(t, err)

	denomTrace = transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(channFeeabsGaia.PortID, channFeeabsGaia.ChannelID, gaia.Config().Denom))
	// uatomOnFeeabs := denomTrace.IBCDenom()

	current_directory, _ := os.Getwd()
	param_change_path := path.Join(current_directory, "proposal", "proposal.json")

	changeParamProposal, err := paramsutils.ParseParamChangeProposalJSON(feeabs.Config().EncodingConfig.Amino, param_change_path)
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

	_, err = cosmos.PollForProposalStatus(ctx, feeabs, height, height+10, "2", cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	_, err = feeabsCli.QueryAllHostZoneConfig(feeabs, ctx)
	require.NoError(t, err)

	// Wait a few blocks for relayer to start and for user accounts to be created
	err = testutil.WaitForBlocks(ctx, 5, feeabs, gaia)
	require.NoError(t, err)

	// Get our Bech32 encoded user addresses
	feeabsUser, gaiaUser := users[0], users[1]

	feeabsUserAddr := sdktypes.MustBech32ifyAddressBytes(feeabs.Config().Bech32Prefix, feeabsUser.Address())
	gaiaUserAddr := sdktypes.MustBech32ifyAddressBytes(gaia.Config().Bech32Prefix, gaiaUser.Address())

	// Compose an IBC transfer and send from Gaia -> Feeabs

	gaiaTokenDenom := transfertypes.GetPrefixedDenom(channFeeabsGaia.ChannelID, channFeeabsGaia.PortID, gaia.Config().Denom)
	gaiaIBCDenom := transfertypes.ParseDenomTrace(gaiaTokenDenom).IBCDenom()

	transferAmount := math.NewInt(1_000)
	transfer := ibc.WalletAmount{
		Address: gaiaUserAddr,
		Denom:   gaia.Config().Denom,
		Amount:  transferAmount,
	}

	gaiaInitialBal, err := gaia.GetBalance(ctx, gaiaUserAddr, gaia.Config().Denom)
	require.NoError(t, err)

	transferTx, err := gaia.SendIBCTransfer(ctx, channGaiaFeeabs.ChannelID, gaiaUser.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	gaiaHeight, err := gaia.Height(ctx)
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, gaia, gaiaHeight, gaiaHeight+10, transferTx.Packet)
	require.NoError(t, err)

	// Assert that the ATOM funds are no longer present in user acc on Gaia and are in the user acc on Feeabs
	feeabsUpdateBal, err := feeabs.GetBalance(ctx, feeabsUserAddr, gaiaIBCDenom)
	require.NoError(t, err)
	require.Equal(t, transferAmount, feeabsUpdateBal)

	gaiaUpdateBal, err := gaia.GetBalance(ctx, gaiaUserAddr, gaia.Config().Denom)
	require.NoError(t, err)
	require.LessOrEqual(t, gaiaInitialBal.Sub(transferAmount), gaiaUpdateBal)

	// Compose an IBC transfer and send from Feeabs -> Gaia
	transferAmount = math.NewInt(1_000)
	ibcFee := sdk.NewCoin(gaiaIBCDenom, sdk.NewInt(1000))
	transfer = ibc.WalletAmount{
		Address: gaiaUserAddr,
		Denom:   gaia.Config().Denom,
		Amount:  transferAmount,
	}

	_, err = SendIBCTransferWithCustomFee(feeabs, ctx, feeabsUser.KeyName(), channFeeabsGaia.ChannelID, transfer, sdk.Coins{ibcFee})
	require.NoError(t, err)

	feeabsHeight, err := feeabs.Height(ctx)
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, feeabs, feeabsHeight, feeabsHeight+10, transferTx.Packet)
	require.NoError(t, err)

	// Get the IBC denom for stake on Gaia
	feeabsTokenDenom := transfertypes.GetPrefixedDenom(channGaiaFeeabs.PortID, channGaiaFeeabs.ChannelID, feeabs.Config().Denom)
	feeabsIBCDenom := transfertypes.ParseDenomTrace(feeabsTokenDenom).IBCDenom()

	// Assert that gaia usre receive the funds from feeabs after the custom fee IBC transfer
	stakeOnGaiaBalance, err := gaia.GetBalance(ctx, gaiaUserAddr, feeabsIBCDenom)
	require.NoError(t, err)

	require.Equal(t, transferAmount, stakeOnGaiaBalance)
}

func SendIBCTransferWithCustomFee(c *cosmos.CosmosChain, ctx context.Context, keyName string, channelID string, amount ibc.WalletAmount, fees sdk.Coins) (string, error) {

	tn := c.Validators[0]
	if len(c.FullNodes) > 0 {
		tn = c.FullNodes[0]
	}
	command := []string{
		"ibc-transfer", "transfer", "transfer", channelID,
		amount.Address, fmt.Sprintf("%s%s", amount.Amount.String(), amount.Denom),
	}
	return tn.ExecTx(ctx, keyName, command...)
}