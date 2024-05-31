package interchaintest

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestStartFeeabs is a basic test to assert that spinning up a Feeabs network with 1 validator works properly.
func TestStartFeeabs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

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

	feeabs := chains[0].(*cosmos.CosmosChain)
	ic := interchaintest.NewInterchain().AddChain(feeabs)
	ctx := context.Background()
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

	a, err := feeabs.AuthQueryModuleAccounts(ctx)
	require.NoError(t, err)
	t.Log("module accounts", a)
}
