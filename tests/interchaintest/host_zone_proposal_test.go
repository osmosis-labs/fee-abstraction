package interchaintest

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	feeabsCli "github.com/osmosis-labs/fee-abstraction/v8/tests/interchaintest/feeabs"
)

func TestHostZoneProposal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	numVals := 1
	numFullNodes := 1

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "feeabs",
			ChainConfig:   feeabsConfig,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	ctx := context.Background()
	feeabs := chains[0].(*cosmos.CosmosChain)
	ic := interchaintest.NewInterchain().AddChain(feeabs)
	client, network := interchaintest.DockerSetup(t)

	require.NoError(t, ic.Build(ctx, nil, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: true,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), genesisWalletAmount, feeabs)
	feeabsUser := users[0]

	ParamChangeProposal(t, ctx, feeabs, feeabsUser, "channel-0", "channel-1", fakeIBCDenom)
	AddHostZoneProposal(t, ctx, feeabs, feeabsUser)

	_, err = feeabsCli.QueryHostZoneConfigWithDenom(feeabs, ctx, fakeIBCDenom)
	require.NoError(t, err)
}
