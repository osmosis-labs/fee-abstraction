package interchaintest

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	feeabsCli "github.com/osmosis-labs/fee-abstraction/v8/tests/interchaintest/feeabs"
	"github.com/osmosis-labs/fee-abstraction/v8/tests/interchaintest/tendermint"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	feeabstypes "github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
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

	channFeeabsOsmosis, channOsmosisFeeabs, channFeeabsGaia, channGaiaFeeabs, channOsmosisGaia, channGaiaOsmosis, channFeeabsOsmosisICQ := channels[0], channels[1], channels[2], channels[3], channels[4], channels[5], channels[6]

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

	// Create pool Osmosis(stake)/uosmo on Osmosis, with 1:1 ratio
	poolID, err := feeabsCli.CreatePool(osmosis, ctx, osmosisUser.KeyName(), cosmos.OsmosisPoolParams{
		Weights:        fmt.Sprintf("5%s,5%s", stakeOnOsmosis, osmosis.Config().Denom),
		InitialDeposit: fmt.Sprintf("95000000%s,95000000%s", stakeOnOsmosis, osmosis.Config().Denom),
		SwapFee:        "0.01",
		ExitFee:        "0",
		FutureGovernor: "",
	})
	require.NoError(t, err)
	require.Equal(t, poolID, "1")

	////////////////////////////////////////////////////////////////////////////////////////
	// Setup propose_pfm
	////////////////////////////////////////////////////////////////////////////////////////

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

	////////////////////////////////////////////////////////////////////////////////////////
	// Setup feeabs module & add host zone via proposals
	////////////////////////////////////////////////////////////////////////////////////////

	ParamChangeProposal(t, ctx, feeabs, feeabsUser, &channFeeabsOsmosis, &channFeeabsOsmosisICQ, stakeOnOsmosis)
	AddHostZoneProposal(t, ctx, feeabs, feeabsUser)
	_, err = feeabsCli.QueryAllHostZoneConfig(feeabs, ctx)
	require.NoError(t, err)

	////////////////////////////////////////////////////////////////////////////////////////
	// Test IBC transfer with custom fee
	////////////////////////////////////////////////////////////////////////////////////////

	// Wait a few blocks for relayer to start and for user accounts to be created
	err = testutil.WaitForBlocks(ctx, 5, feeabs, gaia)
	require.NoError(t, err)

	// Get our Bech32 encoded user addresses
	feeabsUser, gaiaUser := users[0], users[1]

	feeabsUserAddr := sdktypes.MustBech32ifyAddressBytes(feeabs.Config().Bech32Prefix, feeabsUser.Address())
	gaiaUserAddr := sdktypes.MustBech32ifyAddressBytes(gaia.Config().Bech32Prefix, gaiaUser.Address())

	// Compose an IBC transfer and send from Gaia -> Feeabs
	osmoTokenDenom := transfertypes.GetPrefixedDenom(channFeeabsOsmosis.PortID, channFeeabsOsmosis.ChannelID, osmosis.Config().Denom)
	osmoIBCDenom := transfertypes.ParseDenomTrace(osmoTokenDenom).IBCDenom()
	fmt.Println("osmoIBCDenom", osmoIBCDenom)

	transferAmount := math.NewInt(1_000)
	transfer := ibc.WalletAmount{
		Address: feeabsUserAddr,
		Denom:   osmosis.Config().Denom,
		Amount:  transferAmount,
	}

	// Compose an IBC transfer and send from Osmo -> Feeabs
	feeabsInitialBal, err := feeabs.GetBalance(ctx, feeabsUserAddr, osmoIBCDenom)
	require.NoError(t, err)
	osmoInitialBal, err := osmosis.GetBalance(ctx, osmosisUser.FormattedAddress(), osmosis.Config().Denom)
	require.NoError(t, err)

	transferTx, err := osmosis.SendIBCTransfer(ctx, channOsmosisFeeabs.ChannelID, osmosisUser.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	osmosisHeight, err := osmosis.Height(ctx)
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, osmosis, osmosisHeight, osmosisHeight+10, transferTx.Packet)
	require.NoError(t, err)

	// Assert that the OSMO funds are deducted in user acc on Gaia and are in the user acc on Feeabs
	feeabsUpdateBal, err := feeabs.GetBalance(ctx, feeabsUserAddr, osmoIBCDenom)
	require.NoError(t, err)
	require.Equal(t, feeabsInitialBal.Add(transferAmount), feeabsUpdateBal)

	osmoUpdateBal, err := osmosis.GetBalance(ctx, osmosisUser.FormattedAddress(), osmosis.Config().Denom)
	require.NoError(t, err)
	require.GreaterOrEqual(t, osmoInitialBal.Sub(transferAmount).Int64(), osmoUpdateBal.Int64())

	// Fund the feeabs module account with stake in order to pay native fee
	feeabsModuleAddr, err := feeabs.AuthQueryModuleAddress(ctx, feeabstypes.ModuleName)
	fmt.Println("feeabsModuleAddr", feeabsModuleAddr)
	require.NoError(t, err)
	require.NotNil(t, feeabsModuleAddr)
	transfer = ibc.WalletAmount{
		Address: feeabsModuleAddr,
		Denom:   feeabs.Config().Denom,
		Amount:  transferAmount.Mul(math.NewInt(2)),
	}
	err = feeabs.SendFunds(ctx, feeabsUser.KeyName(), transfer)
	require.NoError(t, err)

	// Compose an IBC transfer and send from Feeabs -> Gaia
	transferAmount = math.NewInt(1_000)
	ibcFee := sdk.NewCoin(osmoIBCDenom, math.NewInt(1000))
	transfer = ibc.WalletAmount{
		Address: gaiaUserAddr,
		Denom:   feeabs.Config().Denom,
		Amount:  transferAmount,
	}

	fmt.Println("Module accounts on Feeabs", GetModuleAccounts(feeabs))
	customTransferTx, err := SendIBCTransferWithCustomFee(feeabs, ctx, feeabsUser.KeyName(), channFeeabsGaia.ChannelID, transfer, sdk.Coins{ibcFee})
	require.NoError(t, err)

	feeabsHeight, err := feeabs.Height(ctx)
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, feeabs, feeabsHeight, feeabsHeight+20, customTransferTx.Packet)
	require.NoError(t, err)

	// Get the IBC denom for stake on Gaia
	feeabsTokenDenom := transfertypes.GetPrefixedDenom(channGaiaFeeabs.PortID, channGaiaFeeabs.ChannelID, feeabs.Config().Denom)
	feeabsIBCDenom := transfertypes.ParseDenomTrace(feeabsTokenDenom).IBCDenom()

	// Assert that gaia usre receive the funds from feeabs after the custom fee IBC transfer
	stakeOnGaiaBalance, err := gaia.GetBalance(ctx, gaiaUserAddr, feeabsIBCDenom)
	require.NoError(t, err)

	require.Equal(t, transferAmount, stakeOnGaiaBalance)

	// Compose an IBC transfer and send from Feeabs -> Gaia, with insufficient fee, should fail
	ibcFee = sdk.NewCoin(osmoIBCDenom, math.OneInt())
	transfer = ibc.WalletAmount{
		Address: gaiaUserAddr,
		Denom:   feeabs.Config().Denom,
		Amount:  transferAmount,
	}

	// Compose an IBC transfer and send from Feeabs -> Gaia, with insufficient fee, should fail
	customTransferTx, err = SendIBCTransferWithCustomFee(feeabs, ctx, feeabsUser.KeyName(), channFeeabsGaia.ChannelID, transfer, sdk.Coins{ibcFee})
	require.Error(t, err)

}
func GetModuleAccounts(c *cosmos.CosmosChain) []authtypes.ModuleAccount {
	acc, err := c.AuthQueryModuleAccounts(context.Background())
	if err != nil {
		panic(err)
	}
	return acc
}

func SendIBCTransferWithCustomFee(c *cosmos.CosmosChain, ctx context.Context, keyName string, channelID string, amount ibc.WalletAmount, fees sdk.Coins) (ibc.Tx, error) {
	tn := c.Validators[0]
	if len(c.FullNodes) > 0 {
		tn = c.FullNodes[0]
	}
	command := []string{
		"ibc-transfer", "transfer", "transfer", channelID,
		amount.Address, fmt.Sprintf("%s%s", amount.Amount.String(), amount.Denom), "--fees", fees.String(),
	}
	var tx ibc.Tx
	txHash, err := tn.ExecTx(ctx, keyName, command...)

	if err != nil {
		return tx, fmt.Errorf("send ibc transfer: %w", err)
	}
	txResp, err := c.GetTransaction(txHash)
	if err != nil {
		return tx, fmt.Errorf("failed to get transaction %s: %w", txHash, err)
	}
	if txResp.Code != 0 {
		return tx, fmt.Errorf("error in transaction (code: %d): %s", txResp.Code, txResp.RawLog)
	}
	tx.Height = txResp.Height
	tx.TxHash = txHash
	// In cosmos, user is charged for entire gas requested, not the actual gas used.
	tx.GasSpent = txResp.GasWanted

	const evType = "send_packet"
	events := txResp.Events

	var (
		seq, _           = tendermint.AttributeValue(events, evType, "packet_sequence")
		srcPort, _       = tendermint.AttributeValue(events, evType, "packet_src_port")
		srcChan, _       = tendermint.AttributeValue(events, evType, "packet_src_channel")
		dstPort, _       = tendermint.AttributeValue(events, evType, "packet_dst_port")
		dstChan, _       = tendermint.AttributeValue(events, evType, "packet_dst_channel")
		timeoutHeight, _ = tendermint.AttributeValue(events, evType, "packet_timeout_height")
		timeoutTs, _     = tendermint.AttributeValue(events, evType, "packet_timeout_timestamp")
		data, _          = tendermint.AttributeValue(events, evType, "packet_data")
	)
	tx.Packet.SourcePort = srcPort
	tx.Packet.SourceChannel = srcChan
	tx.Packet.DestPort = dstPort
	tx.Packet.DestChannel = dstChan
	tx.Packet.TimeoutHeight = timeoutHeight
	tx.Packet.Data = []byte(data)

	seqNum, err := strconv.Atoi(seq)
	if err != nil {
		return tx, fmt.Errorf("invalid packet sequence from events %s: %w", seq, err)
	}
	tx.Packet.Sequence = uint64(seqNum)

	timeoutNano, err := strconv.ParseUint(timeoutTs, 10, 64)
	if err != nil {
		return tx, fmt.Errorf("invalid packet timestamp timeout %s: %w", timeoutTs, err)
	}
	tx.Packet.TimeoutTimestamp = ibc.Nanoseconds(timeoutNano)

	return tx, nil
}
