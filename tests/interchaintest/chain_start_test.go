package interchaintest

import (
	"context"
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"
)

// TestStartFeeabs is a basic test to assert that spinning up a Feeabs network with 1 validator works properly.
func TestStartFeeabs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	// Set up chains, users and channels
	ctx := context.Background()
	chains, _, _ := SetupChain(t, ctx)
	feeabs, _, _ := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)
	a, err := feeabs.AuthQueryModuleAccounts(ctx)
	require.NoError(t, err)
	fmt.Println("module accounts", a)
}
