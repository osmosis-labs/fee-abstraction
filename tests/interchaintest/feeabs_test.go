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
	feeabsCli "github.com/notional-labs/fee-abstraction/tests/interchaintest/feeabs"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
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
	_, err = osmosis.ExecuteContract(ctx, osmosisUser.KeyName(), registryContractAddress, msg)
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
		Weights:        fmt.Sprintf("5%s,5%s", stakeOnOsmosis, uatomOnOsmosis),
		InitialDeposit: fmt.Sprintf("95000000%s,950000000%s", stakeOnOsmosis, uatomOnOsmosis),
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
			ChainID: "feeabs",
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
			ChainID: "gaia",
		},
	}
	res = QuerySmartMsgResponse{}
	err = osmosis.QueryContract(ctx, registryContractAddress, queryMsg, &res)
	require.NoError(t, err)
	// store swaprouter
	swapRouterContractID, err := osmosis.StoreContract(ctx, osmosisUser.KeyName(), "./bytecode/swaprouter.wasm")
	require.NoError(t, err)
	// instantiate
	swapRouterContractAddress, err := osmosis.InstantiateContract(ctx, osmosisUser.KeyName(), swapRouterContractID, initMsg, true)
	require.NoError(t, err)

	// execute
	msg = fmt.Sprintf("{\"set_route\":{\"input_denom\":\"%s\",\"output_denom\":\"%s\",\"pool_route\":[{\"pool_id\":\"%s\",\"token_out_denom\":\"%s\"}]}}",
		uatomOnOsmosis,
		stakeOnOsmosis,
		poolID,
		stakeOnOsmosis,
	)
	_, err = osmosis.ExecuteContract(ctx, osmosisUser.KeyName(), swapRouterContractAddress, msg)
	require.NoError(t, err)

	// store xcs
	xcsContractID, err := osmosis.StoreContract(ctx, osmosisUser.KeyName(), "./bytecode/crosschain_swaps.wasm")
	require.NoError(t, err)
	// instantiate
	initMsg = fmt.Sprintf("{\"swap_contract\":\"%s\",\"governor\": \"%s\"}", swapRouterContractAddress, owner)
	xcsContractAddress, err := osmosis.InstantiateContract(ctx, osmosisUser.KeyName(), xcsContractID, initMsg, true)
	_ = xcsContractAddress
	require.NoError(t, err)
	// Swap Feeabs(uatom) to Osmosis
	// send ibc token to feeabs module account
	gaiaHeight, err := gaia.Height(ctx)
	require.NoError(t, err)

	feeabsModule, err := feeabsCli.QueryModuleAccountBalances(feeabs, ctx)
	require.NoError(t, err)
	dstAddress := feeabsModule.Address
	transfer := ibc.WalletAmount{
		Address: dstAddress,
		Denom:   gaia.Config().Denom,
		Amount:  math.NewInt(1_000_000),
	}

	tx, err := gaia.SendIBCTransfer(ctx, channGaiaFeeabs.ChannelID, gaiaUser.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)
	require.NoError(t, tx.Validate())

	_, err = testutil.PollForAck(ctx, gaia, gaiaHeight, gaiaHeight+30, tx.Packet)
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, 1, feeabs, gaia, osmosis)
	require.NoError(t, err)

	denomTrace = transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(channFeeabsGaia.PortID, channFeeabsGaia.ChannelID, gaia.Config().Denom))
	uatomOnFeeabs := denomTrace.IBCDenom()

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

	_, err = cosmos.PollForProposalStatus(ctx, feeabs, height, height+10, "2", cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	_, err = feeabsCli.QueryHostZoneConfig(feeabs, ctx)
	require.NoError(t, err)
	// xcs
	feeabsHeight, err := feeabs.Height(ctx)
	require.NoError(t, err)

	feeabsModule, err = feeabsCli.QueryModuleAccountBalances(feeabs, ctx)
	require.NoError(t, err)
	fmt.Printf("Module Account Balances before swap: %v\n", feeabsModule.Balances)

	transferTx, err := feeabsCli.CrossChainSwap(feeabs, ctx, feeabsUser.KeyName(), uatomOnFeeabs)
	require.NoError(t, err)
	_, err = testutil.PollForAck(ctx, feeabs, feeabsHeight, feeabsHeight+25, transferTx.Packet)
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, 50, feeabs, gaia, osmosis)
	require.NoError(t, err)

	feeabsModule, err = feeabsCli.QueryModuleAccountBalances(feeabs, ctx)
	require.NoError(t, err)
	fmt.Printf("Module Account Balances after swap: %v\n", feeabsModule.Balances)

	balance, err := feeabs.GetBalance(ctx, feeabsModule.Address, feeabs.Config().Denom)
	require.NoError(t, err)
	require.True(t, balance.GT(math.NewInt(1)))
}
