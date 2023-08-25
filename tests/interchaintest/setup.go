package interchaintest

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/icza/dyno"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos/wasm"

	feeabstype "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"

	// "github.com/strangelove-ventures/interchaintest/v7/chain/cosmos/wasm"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

type QueryFeeabsModuleBalacesResponse struct {
	Balances types.Coins
	Address  string
}

type QueryHostChainConfigRespone struct {
	HostChainConfig cosmos.HostChainFeeAbsConfig `protobuf:"bytes,1,opt,name=host_chain_config,json=hostChainConfig,proto3" json:"host_chain_config" yaml:"host_chain_config"`
}

type HasPacketForwarding struct {
	ChainID string `json:"chain"`
}

type QuerySmartMsg struct {
	Packet HasPacketForwarding `json:"has_packet_forwarding"`
}

type QuerySmartMsgResponse struct {
	Response bool `json:"data"`
}

const (
	votingPeriod     = "10s"
	maxDepositPeriod = "10s"
)

var (
	FFeeabsMainRepo  = "osmolabs/fee-abstraction"
	FeeabsICTestRepo = "osmolabs/fee-abstraction-ictest"

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
		GasPrices:           "0.0stake",
		GasAdjustment:       1.1,
		TrustingPeriod:      "112h",
		NoHostMount:         false,
		ModifyGenesis:       modifyGenesisShortProposals(votingPeriod, maxDepositPeriod),
		ConfigFileOverrides: nil,
		EncodingConfig:      feeabsEncoding(),
	}

	pathFeeabsGaia      = "feeabs-gaia"
	pathFeeabsOsmosis   = "feeabs-osmosis"
	pathOsmosisGaia     = "osmosis-gaia"
	genesisWalletAmount = int64(10_000_000)
)

// feeabsEncoding registers the feeabs specific module codecs so that the associated types and msgs
// will be supported when writing to the blocksdb sqlite database.
func feeabsEncoding() *testutil.TestEncodingConfig {
	cfg := wasm.WasmEncoding()

	// register custom types
	feeabstype.RegisterInterfaces(cfg.InterfaceRegistry)

	return cfg
}

func osmosisEncoding() *testutil.TestEncodingConfig {
	cfg := wasm.WasmEncoding()

	// gammtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	// balancertypes.RegisterInterfaces(cfg.InterfaceRegistry)

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

func modifyGenesisShortProposals(votingPeriod string, maxDepositPeriod string) func(ibc.ChainConfig, []byte) ([]byte, error) {
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
		out, err := json.Marshal(g)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
		}
		return out, nil
	}
}
