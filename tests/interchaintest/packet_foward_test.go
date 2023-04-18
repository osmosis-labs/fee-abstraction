package interchaintest

import (
	"context"
	"fmt"
	"testing"
	"time"

	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type PacketMetadata struct {
	Forward *ForwardMetadata `json:"forward"`
}

type ForwardMetadata struct {
	Receiver       string        `json:"receiver"`
	Port           string        `json:"port"`
	Channel        string        `json:"channel"`
	Timeout        time.Duration `json:"timeout"`
	Retries        *uint8        `json:"retries,omitempty"`
	Next           *string       `json:"next,omitempty"`
	RefundSequence *uint64       `json:"refund_sequence,omitempty"`
}

func TestPacketForwardMiddleware(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	client, network := interchaintest.DockerSetup(t)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()

	// Create chain factory with Feeabs and Gaia
	numVals := 1
	numFullNodes := 1
	gasAdjustment := 2.0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "feeabs",
			ChainConfig:   feeabsConfig,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:    "gaia",
			Version: "v9.0.2",
			ChainConfig: ibc.ChainConfig{
				GasPrices: "0.0uatom",
			},
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:    "osmosis",
			Version: "v15.0.0",
			ChainConfig: ibc.ChainConfig{
				GasPrices:      "0.005uosmo",
				EncodingConfig: osmosisEncoding(),
			},
			GasAdjustment: &gasAdjustment,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	feeabs, gaia, osmosis := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)

	r := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		// relayer.CustomDockerImage("ghcr.io/cosmos/relayer", "main", rly.RlyDefaultUidGid),
	).Build(t, client, network)

	ic := interchaintest.NewInterchain().
		AddChain(feeabs).
		AddChain(gaia).
		AddChain(osmosis).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  feeabs,
			Chain2:  gaia,
			Relayer: r,
			Path:    pathFeeabsGaia,
		}).
		AddLink(interchaintest.InterchainLink{
			Chain1:  feeabs,
			Chain2:  osmosis,
			Relayer: r,
			Path:    pathFeeabsOsmosis,
		}).
		AddLink(interchaintest.InterchainLink{
			Chain1:  osmosis,
			Chain2:  gaia,
			Relayer: r,
			Path:    pathOsmosisGaia,
		})

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: true,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	const userFunds = int64(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, feeabs, gaia, osmosis)

	// rly feeabs-osmo
	// Generate new path
	err = r.GeneratePath(ctx, eRep, feeabs.Config().ChainID, osmosis.Config().ChainID, pathFeeabsOsmosis)
	require.NoError(t, err)
	// Create client
	err = r.CreateClients(ctx, eRep, pathFeeabsOsmosis, ibc.DefaultClientOpts())
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, feeabs, osmosis)
	require.NoError(t, err)

	// Create connection
	err = r.CreateConnections(ctx, eRep, pathFeeabsOsmosis)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, feeabs, osmosis)
	require.NoError(t, err)
	// Create channel
	err = r.CreateChannel(ctx, eRep, pathFeeabsOsmosis, ibc.CreateChannelOptions{
		SourcePortName: "transfer",
		DestPortName:   "transfer",
		Order:          ibc.Unordered,
		Version:        "ics20-1",
	})
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, feeabs, osmosis)
	require.NoError(t, err)

	channsFeeabs, err := r.GetChannels(ctx, eRep, feeabs.Config().ChainID)
	require.NoError(t, err)

	channsOsmosis, err := r.GetChannels(ctx, eRep, osmosis.Config().ChainID)
	require.NoError(t, err)

	require.Len(t, channsFeeabs, 1)
	require.Len(t, channsOsmosis, 1)

	channFeeabsOsmosis := channsFeeabs[0]
	require.NotEmpty(t, channFeeabsOsmosis.ChannelID)
	channOsmosisFeeabs := channsOsmosis[0]
	require.NotEmpty(t, channOsmosisFeeabs.ChannelID)
	// rly feeabs-gaia
	// Generate new path
	err = r.GeneratePath(ctx, eRep, feeabs.Config().ChainID, gaia.Config().ChainID, pathFeeabsGaia)
	require.NoError(t, err)
	// Create clients
	err = r.CreateClients(ctx, eRep, pathFeeabsGaia, ibc.DefaultClientOpts())
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, feeabs, gaia)
	require.NoError(t, err)

	// Create connection
	err = r.CreateConnections(ctx, eRep, pathFeeabsGaia)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, feeabs, gaia)
	require.NoError(t, err)

	//Create channel
	err = r.CreateChannel(ctx, eRep, pathFeeabsGaia, ibc.CreateChannelOptions{
		SourcePortName: "transfer",
		DestPortName:   "transfer",
		Order:          ibc.Unordered,
		Version:        "ics20-1",
	})
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, feeabs, gaia)
	require.NoError(t, err)

	channsFeeabs, err = r.GetChannels(ctx, eRep, feeabs.Config().ChainID)
	require.NoError(t, err)

	channsGaia, err := r.GetChannels(ctx, eRep, gaia.Config().ChainID)
	require.NoError(t, err)

	require.Len(t, channsFeeabs, 2)
	require.Len(t, channsGaia, 1)

	var channFeeabsGaia ibc.ChannelOutput
	for _, chann := range channsFeeabs {
		if chann.ChannelID != channFeeabsOsmosis.ChannelID {
			channFeeabsGaia = chann
		}
	}
	require.NotEmpty(t, channFeeabsGaia.ChannelID)

	channGaiaFeeabs := channsGaia[0]
	require.NotEmpty(t, channGaiaFeeabs.ChannelID)
	fmt.Println("-----------------------------------")
	fmt.Printf("channFeeabsOsmosis: %s\n", channFeeabsOsmosis.ChannelID)
	fmt.Printf("channFeeabsGaia: %s\n", channFeeabsGaia.ChannelID)
	fmt.Printf("channGaiaFeeabs: %s - %s\n", channGaiaFeeabs.ChannelID, channGaiaFeeabs.Counterparty.ChannelID)
	fmt.Println("-----------------------------------")
	//rly osmo-gaia
	// Generate new path
	err = r.GeneratePath(ctx, eRep, osmosis.Config().ChainID, gaia.Config().ChainID, pathOsmosisGaia)
	require.NoError(t, err)
	// Create clients
	err = r.CreateClients(ctx, eRep, pathOsmosisGaia, ibc.DefaultClientOpts())
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, osmosis, gaia)
	require.NoError(t, err)
	// Create connection
	err = r.CreateConnections(ctx, eRep, pathOsmosisGaia)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, osmosis, gaia)
	require.NoError(t, err)
	// Create channel
	err = r.CreateChannel(ctx, eRep, pathOsmosisGaia, ibc.CreateChannelOptions{
		SourcePortName: "transfer",
		DestPortName:   "transfer",
		Order:          ibc.Unordered,
		Version:        "ics20-1",
	})
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, osmosis, gaia)
	require.NoError(t, err)

	channsOsmosis, err = r.GetChannels(ctx, eRep, osmosis.Config().ChainID)
	require.NoError(t, err)

	channsGaia, err = r.GetChannels(ctx, eRep, gaia.Config().ChainID)
	require.NoError(t, err)

	require.Len(t, channsOsmosis, 2)
	require.Len(t, channsGaia, 2)

	var channOsmosisGaia ibc.ChannelOutput
	var channGaiaOsmosis ibc.ChannelOutput

	for _, chann := range channsOsmosis {
		if chann.ChannelID != channOsmosisFeeabs.ChannelID {
			channOsmosisGaia = chann
		}
	}
	require.NotEmpty(t, channOsmosisGaia)

	for _, chann := range channsGaia {
		if chann.ChannelID != channGaiaFeeabs.ChannelID {
			channGaiaOsmosis = chann
		}
	}
	require.NotEmpty(t, channGaiaOsmosis)

	// Start the relayer on both paths
	err = r.StartRelayer(ctx, eRep, pathFeeabsGaia, pathFeeabsOsmosis, pathOsmosisGaia)
	require.NoError(t, err)

	t.Cleanup(
		func() {
			err := r.StopRelayer(ctx, eRep)
			if err != nil {
				t.Logf("an error occured while stopping the relayer: %s", err)
			}
		},
	)

	// Get original account balances
	feeabsUser, gaiaUser, osmosisUser := users[0], users[1], users[2]
	_ = feeabsUser
	_ = gaiaUser
	_ = osmosisUser

	const amountToSend = int64(1_000_000_000)

	t.Run("xcs", func(t *testing.T) {
		// Send Gaia uatom to Osmosis
		gaiaUserBalance, err := gaia.GetBalance(ctx, gaiaUser.Bech32Address(gaia.Config().Bech32Prefix), gaia.Config().Denom)
		require.NoError(t, err)
		require.Equal(t, userFunds, gaiaUserBalance)
		dstAddress := osmosisUser.Bech32Address(osmosis.Config().Bech32Prefix)
		transfer := ibc.WalletAmount{
			Address: dstAddress,
			Denom:   gaia.Config().Denom,
			Amount:  amountToSend,
		}

		tx, err := gaia.SendIBCTransfer(ctx, channGaiaOsmosis.ChannelID, gaiaUser.KeyName, transfer, ibc.TransferOptions{})
		require.NoError(t, err)
		require.NoError(t, tx.Validate())

		require.NoError(t, r.FlushPackets(ctx, eRep, pathOsmosisGaia, channOsmosisGaia.ChannelID))
		require.NoError(t, r.FlushAcknowledgements(ctx, eRep, pathOsmosisGaia, channGaiaOsmosis.ChannelID))

		// Send Gaia uatom to Feeabs
		dstAddress = feeabsUser.Bech32Address(feeabs.Config().Bech32Prefix)
		transfer = ibc.WalletAmount{
			Address: dstAddress,
			Denom:   gaia.Config().Denom,
			Amount:  amountToSend,
		}

		tx, err = gaia.SendIBCTransfer(ctx, channGaiaFeeabs.ChannelID, gaiaUser.KeyName, transfer, ibc.TransferOptions{})
		require.NoError(t, err)
		require.NoError(t, tx.Validate())

		require.NoError(t, r.FlushPackets(ctx, eRep, pathFeeabsGaia, channFeeabsGaia.ChannelID))
		require.NoError(t, r.FlushAcknowledgements(ctx, eRep, pathFeeabsGaia, channFeeabsGaia.ChannelID))
		// Send Feeabs stake to Osmosis
		dstAddress = osmosisUser.Bech32Address(osmosis.Config().Bech32Prefix)
		transfer = ibc.WalletAmount{
			Address: dstAddress,
			Denom:   feeabs.Config().Denom,
			Amount:  amountToSend,
		}

		tx, err = feeabs.SendIBCTransfer(ctx, channFeeabsOsmosis.ChannelID, feeabsUser.KeyName, transfer, ibc.TransferOptions{})
		require.NoError(t, err)
		require.NoError(t, tx.Validate())

		require.NoError(t, r.FlushPackets(ctx, eRep, pathFeeabsOsmosis, channOsmosisFeeabs.ChannelID))
		require.NoError(t, r.FlushAcknowledgements(ctx, eRep, pathFeeabsOsmosis, channFeeabsOsmosis.ChannelID))
		// Create pool Osmosis(uatom)/Osmosis(stake) on Osmosis
		denomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(channOsmosisGaia.PortID, channOsmosisGaia.ChannelID, gaia.Config().Denom))
		uatomOnOsmosis := denomTrace.IBCDenom()
		osmosisUserBalance, err := osmosis.GetBalance(ctx, osmosisUser.Bech32Address(osmosis.Config().Bech32Prefix), uatomOnOsmosis)
		require.NoError(t, err)
		require.Equal(t, amountToSend, osmosisUserBalance)

		denomTrace = transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(channOsmosisFeeabs.PortID, channOsmosisFeeabs.ChannelID, feeabs.Config().Denom))
		stakeOnOsmosis := denomTrace.IBCDenom()
		osmosisUserBalance, err = osmosis.GetBalance(ctx, osmosisUser.Bech32Address(osmosis.Config().Bech32Prefix), stakeOnOsmosis)
		require.NoError(t, err)
		require.Equal(t, amountToSend, osmosisUserBalance)

		poolID, err := cosmos.OsmosisCreatePool(osmosis, ctx, osmosisUser.KeyName, cosmos.OsmosisPoolParams{
			Weights:        fmt.Sprintf("5%s,5%s", stakeOnOsmosis, uatomOnOsmosis),
			InitialDeposit: fmt.Sprintf("100000000%s,100000000%s", stakeOnOsmosis, uatomOnOsmosis),
			SwapFee:        "0.01",
			ExitFee:        "0",
			FutureGovernor: "",
		})
		require.NoError(t, err)
		require.Equal(t, poolID, "1")
		// Setup contract on Osmosis
		// Store code crosschain Registry
		crossChainRegistryContractID, err := osmosis.StoreContract(ctx, osmosisUser.KeyName, "./bytecode/crosschain_registry.wasm")
		require.NoError(t, err)
		_ = crossChainRegistryContractID
		// // Instatiate
		owner := osmosisUser.Bech32Address(osmosis.Config().Bech32Prefix)
		swapRegistryInitMsg := fmt.Sprintf("{\"owner\":\"%s\"}", owner)
		address, err := osmosis.InstantiateContract(ctx, osmosisUser.KeyName, crossChainRegistryContractID, swapRegistryInitMsg, true)
		fmt.Printf("xcrAddress: %s\n", address)
		require.NoError(t, err)
		// Execute
		msg := fmt.Sprintf("{\"modify_chain_channel_links\": {\"operations\": [{\"operation\": \"set\",\"source_chain\": \"feeabs\",\"destination_chain\": \"osmosis\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"osmosis\",\"destination_chain\": \"feeabs\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"feeabs\",\"destination_chain\": \"gaia\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"gaia\",\"destination_chain\": \"feeabs\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"osmosis\",\"destination_chain\": \"gaia\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"gaia\",\"destination_chain\": \"osmosis\",\"channel_id\": \"%s\"}]}}",
			channFeeabsOsmosis.ChannelID,
			channOsmosisFeeabs.ChannelID,
			channFeeabsGaia.ChannelID,
			channGaiaFeeabs.ChannelID,
			channOsmosisGaia.ChannelID,
			channGaiaOsmosis.ChannelID)
		txHash, err := osmosis.ExecuteContract(ctx, osmosisUser.KeyName, address, msg)
		require.NoError(t, err)
		_ = txHash
		// txs, _ := osmosis.GetTransaction(txHash)
		// fmt.Printf("txs----------------: %v", txs)
		// Swap Feeabs(uatom) to Osmosis
	})
}
