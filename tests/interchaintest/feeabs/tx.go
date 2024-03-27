package feeabs

import (
	"context"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

func QueryOsmosisTWAP(c *cosmos.CosmosChain, ctx context.Context, keyName string) (string, error) {
	tn := c.Validators[0]
	if len(c.FullNodes) > 0 {
		tn = c.FullNodes[0]
	}
	cmd := []string{"feeabs", "query-osmosis-twap"}
	return tn.ExecTx(ctx, keyName, cmd...)
}
