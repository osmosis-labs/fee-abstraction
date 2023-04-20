package interchaintest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
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

	fmt.Println("-----------------------------------")
	fmt.Printf("channFeeabsOsmosis: %s - %s\n", channFeeabsOsmosis.ChannelID, channFeeabsOsmosis.Counterparty.ChannelID)
	fmt.Printf("channOsmosisFeeabs: %s - %s\n", channOsmosisFeeabs.ChannelID, channOsmosisFeeabs.Counterparty.ChannelID)
	fmt.Printf("channFeeabsGaia: %s - %s\n", channFeeabsGaia.ChannelID, channFeeabsGaia.Counterparty.ChannelID)
	fmt.Printf("channGaiaFeeabs: %s - %s\n", channGaiaFeeabs.ChannelID, channGaiaFeeabs.Counterparty.ChannelID)
	fmt.Printf("channOsmosisGaia: %s - %s\n", channOsmosisGaia.ChannelID, channOsmosisGaia.Counterparty.ChannelID)
	fmt.Printf("channGaiaOsmosis: %s - %s\n", channGaiaOsmosis.ChannelID, channGaiaOsmosis.Counterparty.ChannelID)
	fmt.Println("-----------------------------------")

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
		gaiaHeight, err := gaia.Height(ctx)
		require.NoError(t, err)
		dstAddress := osmosisUser.Bech32Address(osmosis.Config().Bech32Prefix)
		transfer := ibc.WalletAmount{
			Address: dstAddress,
			Denom:   gaia.Config().Denom,
			Amount:  amountToSend,
		}

		tx, err := gaia.SendIBCTransfer(ctx, channGaiaOsmosis.ChannelID, gaiaUser.KeyName, transfer, ibc.TransferOptions{})
		require.NoError(t, err)
		require.NoError(t, tx.Validate())

		_, err = testutil.PollForAck(ctx, gaia, gaiaHeight, gaiaHeight+30, tx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, 1, feeabs, gaia, osmosis)
		require.NoError(t, err)

		// Send Feeabs stake to Osmosis
		feeabsHeight, err := feeabs.Height(ctx)
		require.NoError(t, err)
		dstAddress = osmosisUser.Bech32Address(osmosis.Config().Bech32Prefix)
		transfer = ibc.WalletAmount{
			Address: dstAddress,
			Denom:   feeabs.Config().Denom,
			Amount:  amountToSend,
		}

		tx, err = feeabs.SendIBCTransfer(ctx, channFeeabsOsmosis.ChannelID, feeabsUser.KeyName, transfer, ibc.TransferOptions{})
		require.NoError(t, err)
		require.NoError(t, tx.Validate())

		_, err = testutil.PollForAck(ctx, feeabs, feeabsHeight, feeabsHeight+30, tx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, 1, feeabs, gaia, osmosis)
		require.NoError(t, err)

		// Send Gaia uatom to Feeabs
		gaiaHeight, err = gaia.Height(ctx)
		require.NoError(t, err)
		dstAddress = feeabsUser.Bech32Address(feeabs.Config().Bech32Prefix)
		transfer = ibc.WalletAmount{
			Address: dstAddress,
			Denom:   gaia.Config().Denom,
			Amount:  amountToSend,
		}

		tx, err = gaia.SendIBCTransfer(ctx, channGaiaFeeabs.ChannelID, gaiaUser.KeyName, transfer, ibc.TransferOptions{})
		require.NoError(t, err)
		require.NoError(t, tx.Validate())

		_, err = testutil.PollForAck(ctx, gaia, gaiaHeight, gaiaHeight+30, tx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, 1, feeabs, gaia, osmosis)
		require.NoError(t, err)
		// Setup contract on Osmosis
		// Store code crosschain Registry
		crossChainRegistryContractID, err := osmosis.StoreContract(ctx, osmosisUser.KeyName, "./bytecode/crosschain_registry.wasm")
		require.NoError(t, err)
		_ = crossChainRegistryContractID
		// // Instatiate
		owner := osmosisUser.Bech32Address(osmosis.Config().Bech32Prefix)
		initMsg := fmt.Sprintf("{\"owner\":\"%s\"}", owner)
		registryContractAddress, err := osmosis.InstantiateContract(ctx, osmosisUser.KeyName, crossChainRegistryContractID, initMsg, true)
		require.NoError(t, err)
		// Execute
		msg := fmt.Sprintf("{\"modify_chain_channel_links\": {\"operations\": [{\"operation\": \"set\",\"source_chain\": \"feeabs\",\"destination_chain\": \"osmosis\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"osmosis\",\"destination_chain\": \"feeabs\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"feeabs\",\"destination_chain\": \"gaia\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"gaia\",\"destination_chain\": \"feeabs\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"osmosis\",\"destination_chain\": \"gaia\",\"channel_id\": \"%s\"},{\"operation\": \"set\",\"source_chain\": \"gaia\",\"destination_chain\": \"osmosis\",\"channel_id\": \"%s\"}]}}",
			channFeeabsOsmosis.ChannelID,
			channOsmosisFeeabs.ChannelID,
			channFeeabsGaia.ChannelID,
			channGaiaFeeabs.ChannelID,
			channOsmosisGaia.ChannelID,
			channGaiaOsmosis.ChannelID)
		_, err = osmosis.ExecuteContract(ctx, osmosisUser.KeyName, registryContractAddress, msg)
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
		_, err = osmosis.ExecuteContract(ctx, osmosisUser.KeyName, registryContractAddress, msg)
		require.NoError(t, err)

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
			InitialDeposit: fmt.Sprintf("1000000000%s,1000000000%s", stakeOnOsmosis, uatomOnOsmosis),
			SwapFee:        "0.01",
			ExitFee:        "0",
			FutureGovernor: "",
		})
		require.NoError(t, err)
		require.Equal(t, poolID, "1")

		// store swaprouter
		swapRouterContractID, err := osmosis.StoreContract(ctx, osmosisUser.KeyName, "./bytecode/swaprouter.wasm")
		require.NoError(t, err)
		// instantiate
		swapRouterContractAddress, err := osmosis.InstantiateContract(ctx, osmosisUser.KeyName, swapRouterContractID, initMsg, true)
		require.NoError(t, err)

		// execute
		msg = fmt.Sprintf("{\"set_route\":{\"input_denom\":\"%s\",\"output_denom\":\"%s\",\"pool_route\":[{\"pool_id\":\"%s\",\"token_out_denom\":\"%s\"}]}}",
			uatomOnOsmosis,
			stakeOnOsmosis,
			poolID,
			stakeOnOsmosis,
		)
		txHash, err := osmosis.ExecuteContract(ctx, osmosisUser.KeyName, swapRouterContractAddress, msg)
		require.NoError(t, err)
		_ = txHash

		// store xcs
		xcsContractID, err := osmosis.StoreContract(ctx, osmosisUser.KeyName, "./bytecode/crosschain_swaps.wasm")
		require.NoError(t, err)
		// instantiate
		initMsg = fmt.Sprintf("{\"swap_contract\":\"%s\",\"governor\": \"%s\"}", swapRouterContractAddress, owner)
		xcsContractAddress, err := osmosis.InstantiateContract(ctx, osmosisUser.KeyName, xcsContractID, initMsg, true)
		_ = xcsContractAddress
		require.NoError(t, err)
		// Swap Feeabs(uatom) to Osmosis
		// send ibc token to feeabs module account
		gaiaHeight, err = gaia.Height(ctx)
		require.NoError(t, err)
		feeabsModule, err := QueryFeeabsModuleAccountBalances(feeabs, ctx)
		require.NoError(t, err)
		dstAddress = feeabsModule.Address
		transfer = ibc.WalletAmount{
			Address: dstAddress,
			Denom:   gaia.Config().Denom,
			Amount:  1_000_000,
		}

		tx, err = gaia.SendIBCTransfer(ctx, channGaiaFeeabs.ChannelID, gaiaUser.KeyName, transfer, ibc.TransferOptions{})
		require.NoError(t, err)
		require.NoError(t, tx.Validate())

		_, err = testutil.PollForAck(ctx, gaia, gaiaHeight, gaiaHeight+30, tx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, 1, feeabs, gaia, osmosis)
		require.NoError(t, err)

		denomTrace = transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(channFeeabsGaia.PortID, channFeeabsGaia.ChannelID, gaia.Config().Denom))
		uatomOnFeeabs := denomTrace.IBCDenom()

		feeabsModule, err = QueryFeeabsModuleAccountBalances(feeabs, ctx)
		require.NoError(t, err)

		current_directory, _ := os.Getwd()
		param_change_path := path.Join(current_directory, "proposal", "proposal.json")

		paramTx, err := feeabs.ParamChangeProposal(ctx, feeabsUser.KeyName, param_change_path)
		require.NoError(t, err, "error submitting param change proposal tx")

		err = feeabs.VoteOnProposalAllValidators(ctx, paramTx.ProposalID, cosmos.ProposalVoteYes)
		require.NoError(t, err, "failed to submit votes")

		height, _ := feeabs.Height(ctx)
		_, err = cosmos.PollForProposalStatus(ctx, feeabs, height, height+10, paramTx.ProposalID, cosmos.ProposalStatusPassed)
		require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

		_, err = cosmos.FeeabsAddHostZoneProposal(feeabs, ctx, feeabsUser.KeyName, "./proposal/host_zone.json")
		require.NoError(t, err)

		err = feeabs.VoteOnProposalAllValidators(ctx, "2", cosmos.ProposalVoteYes)
		require.NoError(t, err, "failed to submit votes")

		height, _ = feeabs.Height(ctx)
		_, err = cosmos.PollForProposalStatus(ctx, feeabs, height, height+10, "2", cosmos.ProposalStatusPassed)
		require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

		_, err = QueryFeeabsHostZoneConfig(feeabs, ctx)
		require.NoError(t, err)
		// xcs
		feeabsHeight, err = feeabs.Height(ctx)
		require.NoError(t, err)

		transferTx, err := cosmos.FeeabsCrossChainSwap(feeabs, ctx, feeabsUser.KeyName, uatomOnFeeabs)
		require.NoError(t, err)
		_, err = testutil.PollForAck(ctx, feeabs, feeabsHeight, feeabsHeight+100, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, 1, feeabs)
		require.NoError(t, err)

		err = testutil.WaitForBlocks(ctx, 100, feeabs, gaia, osmosis)
		require.NoError(t, err)

		balance, err := feeabs.GetBalance(ctx, feeabsModule.Address, feeabs.Config().Denom)
		require.NoError(t, err)
		require.Greater(t, balance, 1)
	})
}

func QueryFeeabsHostZoneConfig(c *cosmos.CosmosChain, ctx context.Context) (*QueryHostChainConfigRespone, error) {
	cmd := []string{"feeabs", "all-host-chain-config"}
	stdout, _, err := c.ExecQuery(ctx, cmd)
	if err != nil {
		return &QueryHostChainConfigRespone{}, err
	}

	var hostZoneConfig QueryHostChainConfigRespone
	err = json.Unmarshal(stdout, &hostZoneConfig)
	if err != nil {
		return &QueryHostChainConfigRespone{}, err
	}

	return &hostZoneConfig, nil
}

func QueryFeeabsModuleAccountBalances(c *cosmos.CosmosChain, ctx context.Context) (*QueryFeeabsModuleBalacesResponse, error) {
	cmd := []string{"feeabs", "module-balances"}
	stdout, _, err := c.ExecQuery(ctx, cmd)
	if err != nil {
		return &QueryFeeabsModuleBalacesResponse{}, err
	}

	var feeabsModule QueryFeeabsModuleBalacesResponse
	err = json.Unmarshal(stdout, &feeabsModule)
	if err != nil {
		return &QueryFeeabsModuleBalacesResponse{}, err
	}

	return &feeabsModule, nil
}
