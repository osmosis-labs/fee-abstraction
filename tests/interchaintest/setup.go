package interchaintest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/icza/dyno"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos/wasm"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	balancertypes "github.com/osmosis-labs/osmosis/v25/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v25/x/gamm/types"

	feeabstype "github.com/osmosis-labs/fee-abstraction/v8/x/feeabs/types"
)

type HasPacketForwarding struct {
	Chain string `json:"chain"`
}

type QuerySmartMsg struct {
	Packet HasPacketForwarding `json:"has_packet_forwarding"`
}

type QuerySmartMsgResponse struct {
	Data bool `json:"data"`
}

const (
	votingPeriod     = "10s"
	maxDepositPeriod = "10s"
	queryEpochTime   = "10s"
)

var (
	FFeeabsMainRepo     = "osmolabs/fee-abstraction"
	FeeabsICTestRepo    = "osmolabs/fee-abstraction-ictest"
	IBCRelayerImage     = "ghcr.io/cosmos/relayer"
	IBCRelayerVersion   = "latest"
	GaiaImageVersion    = "v14.1.0"
	OsmosisImageVersion = "v22.0.1"

	repo, version = GetDockerImageInfo()

	feeabsImage = ibc.DockerImage{
		Repository: repo,
		Version:    version,
		UidGid:     "1025:1025",
	}

	feeabsConfig = ibc.ChainConfig{
		Type:                "cosmos",
		Name:                "feeabs",
		ChainID:             "feeabs-2",
		Images:              []ibc.DockerImage{feeabsImage},
		Bin:                 "feeappd",
		Bech32Prefix:        "feeabs",
		Denom:               "stake",
		CoinType:            "118",
		GasPrices:           "0.005stake",
		GasAdjustment:       1.5,
		TrustingPeriod:      "112h",
		NoHostMount:         false,
		ModifyGenesis:       modifyGenesisShortProposals(votingPeriod, maxDepositPeriod, queryEpochTime),
		ConfigFileOverrides: nil,
		EncodingConfig:      feeabsEncoding(),
	}

	pathFeeabsGaia      = "feeabs-gaia"
	pathFeeabsOsmosis   = "feeabs-osmosis"
	pathOsmosisGaia     = "osmosis-gaia"
	genesisWalletAmount = math.NewInt(100_000_000_000)
	amountToSend        = math.NewInt(1_000_000_000)
)

// feeabsEncoding registers the feeabs specific module codecs so that the associated types and msgs
// will be supported when writing to the blocksdb sqlite database.
func feeabsEncoding() *moduletestutil.TestEncodingConfig {
	cfg := wasm.WasmEncoding()

	// register custom types
	feeabstype.RegisterInterfaces(cfg.InterfaceRegistry)

	return cfg
}

func osmosisEncoding() *moduletestutil.TestEncodingConfig {
	cfg := wasm.WasmEncoding()

	gammtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	balancertypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return cfg
}

// GetDockerImageInfo returns the appropriate repo and branch version string for integration with the CI pipeline.
// The remote runner sets the BRANCH_CI env var. If present, interchaintest will use the docker image pushed up to the repo.
// If testing locally, user should run `make docker-build-debug` and interchaintest will use the local image.
func GetDockerImageInfo() (repo, version string) {
	branchVersion, found := os.LookupEnv("BRANCH_CI")
	repo = FeeabsICTestRepo
	if !found {
		// make local-image
		repo = "feeapp"
		branchVersion = "debug"
	}

	// github converts / to - for pushed docker images
	branchVersion = strings.ReplaceAll(branchVersion, "/", "-")
	return repo, branchVersion
}

func modifyGenesisWhitelistTwapQueryOsmosis() func(ibc.ChainConfig, []byte) ([]byte, error) {
	return func(chainConfig ibc.ChainConfig, genbz []byte) ([]byte, error) {
		g := make(map[string]interface{})
		if err := json.Unmarshal(genbz, &g); err != nil {
			return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
		}
		// "interchainquery":{"host_port":"icqhost","params":{"allow_queries":[],"host_enabled":true}}
		whitelist := "/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow"
		if err := dyno.Append(g, whitelist, "app_state", "interchainquery", "params", "allow_queries"); err != nil {
			return nil, fmt.Errorf("failed to set whitelist in genesis json: %w", err)
		}
		fmt.Println("Genesis file updated", g)
		out, err := json.Marshal(g)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
		}
		return out, nil
	}
}

func modifyGenesisShortProposals(votingPeriod string, maxDepositPeriod string, queryEpochTime string) func(ibc.ChainConfig, []byte) ([]byte, error) {
	return func(chainConfig ibc.ChainConfig, genbz []byte) ([]byte, error) {
		g := make(map[string]interface{})
		if err := json.Unmarshal(genbz, &g); err != nil {
			return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
		}
		if err := dyno.Set(g, votingPeriod, "app_state", "gov", "params", "voting_period"); err != nil {
			return nil, fmt.Errorf("failed to set voting period in genesis json: %w", err)
		}
		if err := dyno.Set(g, maxDepositPeriod, "app_state", "gov", "params", "max_deposit_period"); err != nil {
			return nil, fmt.Errorf("failed to set voting period in genesis json: %w", err)
		}
		if err := dyno.Set(g, chainConfig.Denom, "app_state", "gov", "params", "min_deposit", 0, "denom"); err != nil {
			return nil, fmt.Errorf("failed to set voting period in genesis json: %w", err)
		}
		if err := dyno.Set(g, queryEpochTime, "app_state", "feeabs", "epochs", 0, "duration"); err != nil {
			return nil, fmt.Errorf("failed to set query epoch time in genesis json: %w", err)
		}
		if err := dyno.Set(g, queryEpochTime, "app_state", "feeabs", "epochs", 1, "duration"); err != nil {
			return nil, fmt.Errorf("failed to set query epoch time in genesis json: %w", err)
		}
		fmt.Println("Genesis file updated", g)
		out, err := json.Marshal(g)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
		}
		return out, nil
	}
}

func SetupChain(t *testing.T, ctx context.Context) ([]ibc.Chain, []ibc.Wallet, []ibc.ChannelOutput) {
	t.Helper()
	client, network := interchaintest.DockerSetup(t)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

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
			Version: GaiaImageVersion,
			ChainConfig: ibc.ChainConfig{
				GasPrices: "0.0uatom",
			},
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:    "osmosis",
			Version: OsmosisImageVersion,
			ChainConfig: ibc.ChainConfig{
				GasPrices:      "0.005uosmo",
				EncodingConfig: osmosisEncoding(),
				ModifyGenesis:  modifyGenesisWhitelistTwapQueryOsmosis(),
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
		relayer.CustomDockerImage(IBCRelayerImage, IBCRelayerVersion, "100:1000"),
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
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: true,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), genesisWalletAmount, feeabs, gaia, osmosis)

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
		SourcePortName: "feeabs",
		DestPortName:   "icqhost",
		Order:          ibc.Unordered,
		Version:        "icq-1",
	})
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, feeabs, osmosis)
	require.NoError(t, err)
	var chanels []ibc.ChannelOutput
	channsFeeabs, err := r.GetChannels(ctx, eRep, feeabs.Config().ChainID)
	require.NoError(t, err)

	channsOsmosis, err := r.GetChannels(ctx, eRep, osmosis.Config().ChainID)
	require.NoError(t, err)

	require.Len(t, channsFeeabs, 1)
	require.Len(t, channsOsmosis, 1)

	chanFeeabsOsmosisFeeapp := channsFeeabs[0]
	channOsmosisFeeabsICQ := channsOsmosis[0]
	err = r.CreateChannel(ctx, eRep, pathFeeabsOsmosis, ibc.CreateChannelOptions{
		SourcePortName: "transfer",
		DestPortName:   "transfer",
		Order:          ibc.Unordered,
		Version:        "ics20-1",
	})
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 5, feeabs, gaia)
	require.NoError(t, err)
	channFeeabsIncludeTransfer, err := r.GetChannels(ctx, eRep, feeabs.Config().ChainID)
	require.NoError(t, err)
	channOsmosisIncludeTransfer, err := r.GetChannels(ctx, eRep, osmosis.Config().ChainID)
	require.NoError(t, err)
	var channFeeabsOsmosisTransfer ibc.ChannelOutput
	var channOsmosisFeeabsTransfer ibc.ChannelOutput
	for _, chann := range channFeeabsIncludeTransfer {
		if chann.ChannelID != chanFeeabsOsmosisFeeapp.ChannelID {
			channFeeabsOsmosisTransfer = chann
		}
	}
	require.NotEmpty(t, channFeeabsOsmosisTransfer)
	for _, chann := range channOsmosisIncludeTransfer {
		if chann.ChannelID != channOsmosisFeeabsICQ.ChannelID {
			channOsmosisFeeabsTransfer = chann
		}
	}

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

	// Create channel
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

	require.Len(t, channsFeeabs, 3)
	require.Len(t, channsGaia, 1)

	var channFeeabsGaia ibc.ChannelOutput
	for _, chann := range channsFeeabs {
		if chann.ChannelID != channFeeabsOsmosisTransfer.ChannelID && chann.ChannelID != chanFeeabsOsmosisFeeapp.ChannelID {
			channFeeabsGaia = chann
		}
	}
	require.NotEmpty(t, channFeeabsGaia.ChannelID)

	channGaiaFeeabs := channsGaia[0]
	require.NotEmpty(t, channGaiaFeeabs.ChannelID)
	// rly osmo-gaia
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

	require.Len(t, channsOsmosis, 3)
	require.Len(t, channsGaia, 2)

	var channOsmosisGaia ibc.ChannelOutput
	var channGaiaOsmosis ibc.ChannelOutput

	for _, chann := range channsOsmosis {
		if chann.ChannelID != channOsmosisFeeabsTransfer.ChannelID && chann.ChannelID != channOsmosisFeeabsICQ.ChannelID {
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
	fmt.Printf("channFeeabsOsmosisTransfer: %s - %s\n", channFeeabsOsmosisTransfer.ChannelID, channFeeabsOsmosisTransfer.Counterparty.ChannelID)
	fmt.Printf("channOsmosisFeeabsTransfer: %s - %s\n", channOsmosisFeeabsTransfer.ChannelID, channOsmosisFeeabsTransfer.Counterparty.ChannelID)
	fmt.Printf("channFeeabsOsmosisfeeabs: %s - %s\n", chanFeeabsOsmosisFeeapp.ChannelID, chanFeeabsOsmosisFeeapp.Counterparty.ChannelID)
	fmt.Printf("channOsmosisFeeabsIcq: %s - %s\n", channOsmosisFeeabsICQ.ChannelID, channOsmosisFeeabsICQ.Counterparty.ChannelID)
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
				t.Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)
	chanels = append(chanels, channFeeabsOsmosisTransfer, channOsmosisFeeabsTransfer, channFeeabsGaia, channGaiaFeeabs, channOsmosisGaia, channGaiaOsmosis, chanFeeabsOsmosisFeeapp, channOsmosisFeeabsICQ)
	feeabsUser, gaiaUser, osmosisUser := users[0], users[1], users[2]

	// Send Gaia uatom to Osmosis
	gaiaHeight, err := gaia.Height(ctx)
	require.NoError(t, err)
	dstAddress := sdktypes.MustBech32ifyAddressBytes(osmosis.Config().Bech32Prefix, osmosisUser.Address())
	transfer := ibc.WalletAmount{
		Address: dstAddress,
		Denom:   gaia.Config().Denom,
		Amount:  amountToSend,
	}

	tx, err := gaia.SendIBCTransfer(ctx, channGaiaOsmosis.ChannelID, gaiaUser.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)
	require.NoError(t, tx.Validate())

	_, err = testutil.PollForAck(ctx, gaia, gaiaHeight, gaiaHeight+30, tx.Packet)
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, 1, feeabs, gaia, osmosis)
	require.NoError(t, err)

	// Send Feeabs stake to Osmosis
	feeabsHeight, err := feeabs.Height(ctx)
	require.NoError(t, err)
	dstAddress = sdktypes.MustBech32ifyAddressBytes(osmosis.Config().Bech32Prefix, osmosisUser.Address())
	transfer = ibc.WalletAmount{
		Address: dstAddress,
		Denom:   feeabs.Config().Denom,
		Amount:  amountToSend,
	}

	tx, err = feeabs.SendIBCTransfer(ctx, channFeeabsOsmosisTransfer.ChannelID, feeabsUser.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)
	require.NoError(t, tx.Validate())

	_, err = testutil.PollForAck(ctx, feeabs, feeabsHeight, feeabsHeight+30, tx.Packet)
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, 1, feeabs, gaia, osmosis)
	require.NoError(t, err)

	// Send Gaia uatom to Feeabs
	gaiaHeight, err = gaia.Height(ctx)
	require.NoError(t, err)
	dstAddress = sdktypes.MustBech32ifyAddressBytes(feeabs.Config().Bech32Prefix, feeabsUser.Address())
	transfer = ibc.WalletAmount{
		Address: dstAddress,
		Denom:   gaia.Config().Denom,
		Amount:  amountToSend,
	}

	tx, err = gaia.SendIBCTransfer(ctx, channGaiaFeeabs.ChannelID, gaiaUser.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)
	require.NoError(t, tx.Validate())

	_, err = testutil.PollForAck(ctx, gaia, gaiaHeight, gaiaHeight+30, tx.Packet)
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, 5, feeabs, gaia, osmosis)
	require.NoError(t, err)

	return chains, users, chanels
}

// SetupOsmosisContracts setup osmosis contracts for crosschain swap.
// There are three main contracts
// 1. crosschain-registry: https://github.com/osmosis-labs/osmosis/blob/main/cosmwasm/contracts/crosschain-swaps/README.md
// 2. swaprouter: https://github.com/osmosis-labs/osmosis/tree/main/cosmwasm/contracts/swaprouter
// 3. crosschain-swaps: https://github.com/osmosis-labs/osmosis/blob/main/cosmwasm/contracts/crosschain-swaps/README.md
func SetupOsmosisContracts(t *testing.T,
	ctx context.Context,
	osmosis *cosmos.CosmosChain,
	user ibc.Wallet,
) ([]string, error) {
	t.Helper()
	registryWasm := "./bytecode/crosschain_registry.wasm"
	swaprouterWasm := "./bytecode/swaprouter.wasm"
	xcsV2Wasm := "./bytecode/crosschain_swaps.wasm"

	// Store crosschain registry contract
	registryCodeId, err := osmosis.StoreContract(ctx, user.KeyName(), registryWasm)
	if err != nil {
		return nil, err
	}
	t.Logf("crosschain registry code id: %s\n", registryCodeId)

	// Store swap router contract
	swaprouterCodeId, err := osmosis.StoreContract(ctx, user.KeyName(), swaprouterWasm)
	if err != nil {
		return nil, err
	}
	t.Logf("swap router code id: %s\n", swaprouterCodeId)

	// Store crosschain swaps contract
	xcsV2CodeId, err := osmosis.StoreContract(ctx, user.KeyName(), xcsV2Wasm)
	if err != nil {
		return nil, err
	}
	t.Logf("crosschain swaps code id: %s\n", xcsV2CodeId)

	// Instantiate contracts
	// 1. Crosschain Registry Contract
	owner := sdktypes.MustBech32ifyAddressBytes(osmosis.Config().Bech32Prefix, user.Address())
	initMsg := fmt.Sprintf("{\"owner\":\"%s\"}", owner)

	registryContractAddr, err := osmosis.InstantiateContract(ctx, user.KeyName(), registryCodeId, initMsg, true)
	if err != nil {
		return nil, err
	}
	t.Logf("registry contract address: %s\n", registryContractAddr)

	// 2. Swap Router Contract
	swaprouterContractAddr, err := osmosis.InstantiateContract(ctx, user.KeyName(), swaprouterCodeId, initMsg, true)
	if err != nil {
		return nil, err
	}
	t.Logf("swap router contract address: %s\n", swaprouterContractAddr)

	// 3. Crosschain Swaps Contract
	initMsg = fmt.Sprintf("{\"swap_contract\":\"%s\",\"governor\": \"%s\",\"registry_contract\": \"%s\"}", swaprouterContractAddr, owner, registryContractAddr)
	xcsV2ContractAddr, err := osmosis.InstantiateContract(ctx, user.KeyName(), xcsV2CodeId, initMsg, true)
	if err != nil {
		return nil, err
	}
	t.Logf("crosschain swaps contract address: %s", xcsV2ContractAddr)

	return []string{registryContractAddr, swaprouterContractAddr, xcsV2ContractAddr}, nil
}
