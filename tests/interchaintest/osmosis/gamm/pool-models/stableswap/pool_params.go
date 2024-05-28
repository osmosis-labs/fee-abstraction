package stableswap

import (
	"github.com/osmosis-labs/fee-abstraction/v7/tests/interchaintest/osmosis/gamm/types"
	"github.com/osmosis-labs/fee-abstraction/v7/tests/interchaintest/osmosis/osmomath"
)

func (params PoolParams) Validate() error {
	if params.ExitFee.IsNegative() {
		return types.ErrNegativeExitFee
	}

	if params.ExitFee.GTE(osmomath.OneDec()) {
		return types.ErrTooMuchExitFee
	}

	if params.SwapFee.IsNegative() {
		return types.ErrNegativeSpreadFactor
	}

	if params.SwapFee.GTE(osmomath.OneDec()) {
		return types.ErrTooMuchSpreadFactor
	}
	return nil
}
