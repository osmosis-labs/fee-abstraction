package interchaintest

import (
	"context"
	"fmt"
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestFeeabsGaiaIBCTransfer spins up a Feeabs and Gaia network, initializes an IBC connection between them,
// and sends an ICS20 token transfer from Feeabs->Gaia and then back from Gaia->Feeabs.
func TestFeeabsGaiaIBCTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

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
			Name:          "gaia",
			Version:       "v9.0.2",
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:          "osmosis",
			Version:       "v15.0.0",
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	feeabs, gaia, osmosis := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)

	// Create relayer factory to utilize the go-relayer
	client, network := interchaintest.DockerSetup(t)

	r1 := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t)).Build(t, client, network)
	r2 := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t)).Build(t, client, network)
	r3 := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t)).Build(t, client, network)
	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic1 := interchaintest.NewInterchain().
		AddChain(feeabs).
		AddChain(gaia).
		AddRelayer(r1, "rly1").
		AddLink(interchaintest.InterchainLink{
			Chain1:  feeabs,
			Chain2:  gaia,
			Relayer: r1,
			Path:    pathFeeabsGaia,
		})
	ic2 := interchaintest.NewInterchain().
		AddChain(feeabs).
		AddChain(osmosis).
		AddRelayer(r2, "rly2").
		AddLink(interchaintest.InterchainLink{
			Chain1:  feeabs,
			Chain2:  osmosis,
			Relayer: r2,
			Path:    pathFeeabsOsmosis,
		})
	ic3 := interchaintest.NewInterchain().
		AddChain(osmosis).
		AddChain(gaia).
		AddRelayer(r3, "rly3").
		AddLink(interchaintest.InterchainLink{
			Chain1:  osmosis,
			Chain2:  gaia,
			Relayer: r3,
			Path:    pathOsmosisGaia,
		})
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	err = ic1.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,

		// This can be used to write to the block database which will index all block data e.g. txs, msgs, events, etc.
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
	})
	require.NoError(t, err)

	err = ic2.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,

		// This can be used to write to the block database which will index all block data e.g. txs, msgs, events, etc.
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
	})
	require.NoError(t, err)

	err = ic3.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,

		// This can be used to write to the block database which will index all block data e.g. txs, msgs, events, etc.
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = ic1.Close()
		_ = ic2.Close()
		_ = ic3.Close()
	})

	// Start the relayer
	require.NoError(t, r1.StartRelayer(ctx, eRep, pathFeeabsGaia))
	t.Cleanup(
		func() {
			err := r1.StopRelayer(ctx, eRep)
			if err != nil {
				panic(fmt.Errorf("an error occurred while stopping the relayer: %s", err))
			}
		},
	)

	require.NoError(t, r2.StartRelayer(ctx, eRep, pathFeeabsOsmosis))
	t.Cleanup(
		func() {
			err := r2.StopRelayer(ctx, eRep)
			if err != nil {
				panic(fmt.Errorf("an error occurred while stopping the relayer: %s", err))
			}
		},
	)

	require.NoError(t, r3.StartRelayer(ctx, eRep, pathOsmosisGaia))
	t.Cleanup(
		func() {
			err := r3.StopRelayer(ctx, eRep)
			if err != nil {
				panic(fmt.Errorf("an error occurred while stopping the relayer: %s", err))
			}
		},
	)

	// Create some user accounts on both chains
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), genesisWalletAmount, feeabs, gaia, osmosis)

	// Wait a few blocks for relayer to start and for user accounts to be created
	err = testutil.WaitForBlocks(ctx, 5, feeabs, gaia, osmosis)
	require.NoError(t, err)

	// Get our Bech32 encoded user addresses
	feeabsUser, gaiaUser, osmosisUser := users[0], users[1], users[2]

	feeabsUserAddr := feeabsUser.Bech32Address(feeabs.Config().Bech32Prefix)
	gaiaUserAddr := gaiaUser.Bech32Address(gaia.Config().Bech32Prefix)
	osmosisUserAddr := osmosisUser.Bech32Address(osmosis.Config().Bech32Prefix)

	// Get original account balances
	feeabsOrigBal, err := feeabs.GetBalance(ctx, feeabsUserAddr, feeabs.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, genesisWalletAmount, feeabsOrigBal)

	gaiaOrigBal, err := gaia.GetBalance(ctx, gaiaUserAddr, gaia.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, genesisWalletAmount, gaiaOrigBal)

	osmosisOrigBal, err := osmosis.GetBalance(ctx, osmosisUserAddr, osmosis.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, genesisWalletAmount, osmosisOrigBal)

	// Compose an IBC transfer and send from feeabs -> Gaia
	const transferAmount = int64(1_000)
	transfer := ibc.WalletAmount{
		Address: gaiaUserAddr,
		Denom:   feeabs.Config().Denom,
		Amount:  transferAmount,
	}

	channel, err := ibc.GetTransferChannel(ctx, r1, eRep, feeabs.Config().ChainID, gaia.Config().ChainID)
	require.NoError(t, err)

	transferTx, err := feeabs.SendIBCTransfer(ctx, channel.ChannelID, feeabsUserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	feeabsHeight, err := feeabs.Height(ctx)
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, feeabs, feeabsHeight, feeabsHeight+10, transferTx.Packet)
	require.NoError(t, err)

	// Get the IBC denom for stake on Gaia
	feeabsTokenDenom := transfertypes.GetPrefixedDenom(channel.Counterparty.PortID, channel.Counterparty.ChannelID, feeabs.Config().Denom)
	feeabsIBCDenom := transfertypes.ParseDenomTrace(feeabsTokenDenom).IBCDenom()

	// Assert that the funds are no longer present in user acc on feeabs and are in the user acc on Gaia
	feeabsUpdateBal, err := feeabs.GetBalance(ctx, feeabsUserAddr, feeabs.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, feeabsOrigBal-transferAmount, feeabsUpdateBal)

	gaiaUpdateBal, err := gaia.GetBalance(ctx, gaiaUserAddr, feeabsIBCDenom)
	require.NoError(t, err)
	require.Equal(t, transferAmount, gaiaUpdateBal)

	// Compose an IBC transfer and send from Gaia -> Feeabs
	transfer = ibc.WalletAmount{
		Address: feeabsUserAddr,
		Denom:   feeabsIBCDenom,
		Amount:  transferAmount,
	}

	transferTx, err = gaia.SendIBCTransfer(ctx, channel.Counterparty.ChannelID, gaiaUserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	gaiaHeight, err := gaia.Height(ctx)
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, gaia, gaiaHeight, gaiaHeight+10, transferTx.Packet)
	require.NoError(t, err)

	// Assert that the funds are now back on feeabs and not on Gaia
	feeabsUpdateBal, err = feeabs.GetBalance(ctx, feeabsUserAddr, feeabs.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, feeabsOrigBal, feeabsUpdateBal)

	gaiaUpdateBal, err = gaia.GetBalance(ctx, gaiaUserAddr, feeabsIBCDenom)
	require.NoError(t, err)
	require.Equal(t, int64(0), gaiaUpdateBal)
}
