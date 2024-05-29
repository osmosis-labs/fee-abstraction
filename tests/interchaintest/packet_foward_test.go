package interchaintest

import (
	"context"
	"fmt"
	"testing"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	feeabsCli "github.com/osmosis-labs/fee-abstraction/v8/tests/interchaintest/feeabs"
)

func TestPacketForwardMiddleware(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	// Set up chains, users and channels
	ctx := context.Background()
	chains, users, channels := SetupChain(t, ctx)
	feeabs, gaia, osmosis := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)

	_, _, osmosisUser := users[0], users[1], users[2]

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
}
