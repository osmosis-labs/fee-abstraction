package interchaintest

import (
	"context"
	"testing"
	"time"

	interchaintest "github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
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
				GasPrices: "0.005uosmo",
			},
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

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	const userFunds = int64(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, feeabs, gaia, osmosis)

	feeabsGaiaChannel, err := ibc.GetTransferChannel(ctx, r, eRep, feeabs.Config().ChainID, gaia.Config().ChainID)
	require.NoError(t, err)
	gaiaFeeabsChannel := feeabsGaiaChannel.Counterparty
	_ = gaiaFeeabsChannel

	feeabsOsmosisChannel, err := ibc.GetTransferChannel(ctx, r, eRep, feeabs.Config().ChainID, osmosis.Config().ChainID)
	require.NoError(t, err)
	osmosisFeeabsChannel := feeabsOsmosisChannel.Counterparty
	_ = osmosisFeeabsChannel

	osmosisGaiaChannel, err := ibc.GetTransferChannel(ctx, r, eRep, osmosis.Config().ChainID, gaia.Config().ChainID)
	require.NoError(t, err)
	gaiaOsmosisChannel := osmosisGaiaChannel.Counterparty
	_ = gaiaOsmosisChannel

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

	// const transferAmount int64 = 100000

	// Compose the prefixed denoms and ibc denom for asserting balances
	// firstHopDenom := transfertypes.GetPrefixedDenom(gaiafeeabsChannel.PortID, gaiafeeabsChannel.ChannelID, feeabs.Config().Denom)
	// secondHopDenom := transfertypes.GetPrefixedDenom(osmosisfeeabsChannel.PortID, osmosisfeeabsChannel.ChannelID, firstHopDenom)
	// thirdHopDenom := transfertypes.GetPrefixedDenom(gaiaosmosisChannel.PortID, gaiaosmosisChannel.ChannelID, secondHopDenom)

	// firstHopDenomTrace := transfertypes.ParseDenomTrace(firstHopDenom)
	// secondHopDenomTrace := transfertypes.ParseDenomTrace(secondHopDenom)
	// thirdHopDenomTrace := transfertypes.ParseDenomTrace(thirdHopDenom)

	// firstHopIBCDenom := firstHopDenomTrace.IBCDenom()
	// secondHopIBCDenom := secondHopDenomTrace.IBCDenom()
	// thirdHopIBCDenom := thirdHopDenomTrace.IBCDenom()

	// firstHopEscrowAccount := transfertypes.GetEscrowAddress(feeabsgaiaChannel.PortID, feeabsgaiaChannel.ChannelID).String()
	// secondHopEscrowAccount := transfertypes.GetEscrowAddress(feeabsosmosisChannel.PortID, feeabsosmosisChannel.ChannelID).String()
	// thirdHopEscrowAccount := transfertypes.GetEscrowAddress(osmosisgaiaChannel.PortID, osmosisgaiaChannel.ChannelID).String()

	t.Run("forward a->b->a", func(t *testing.T) {
		// Setup contract on Osmosis
		// Store code crosschain Registry
		crossChainRegistryContractID, err := feeabs.StoreContract(ctx, feeabsUser.KeyName, "./bytecode/crosschain_registry.wasm")
		require.NoError(t, err)
		_ = crossChainRegistryContractID
		// // Instatiate
		// owner := feeabsUser.Bech32Address(feeabs.Config().Bech32Prefix)
		// swapRegistryInitMsg := fmt.Sprintf("{\"owner\":\"%s\"}", owner)
		// address, err := feeabs.InstantiateContract(ctx, feeabsUser.KeyName, crossChainRegistryContractID, swapRegistryInitMsg, true)
		// require.NoError(t, err)
		// feeabs.FullNodes
		// // Execute
		// msg := fmt.Sprintf("{\"modify_chain_channel_links\": {\"operations\": [{\"operation\": \"set\",\"source_chain\": \"feeabs\",\"destination_chain\": \"osmosis\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"osmosis\",\"destination_chain\": \"feeabs\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"feeabs\",\"destination_chain\": \"gaia\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"gaia\",\"destination_chain\": \"feeabs\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"osmosis\",\"destination_chain\": \"gaia\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"gaia\",\"destination_chain\": \"osmosis\",\"channel_id\": \"%s\"}]}}", feeabsOsmosisChannel, osmosisFeeabsChannel, feeabsGaiaChannel, gaiaFeeabsChannel, osmosisGaiaChannel, gaiaOsmosisChannel)
		// txHash, err := feeabs.ExecuteContract(ctx, feeabsUser.KeyName, address, msg)
		// require.NoError(t, err)
		// fmt.Printf("Hash----------------: %s", txHash)
		// tx, err := feeabs.GetTransaction(txHash)
		// fmt.Printf("tx----------------: %v", tx)

		// Send Gaia uatom to Osmosis

		// Send Gaia uatom to Feeabs

		// Send Feeabs stake to Osmosis

		// Create pool Osmosis(uatom)/Osmosis(stake) on Osmosis

		// Swap Feeabs(uatom) to Osmosis
	})
}
