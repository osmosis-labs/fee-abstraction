package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/fee-abstraction/v6/x/feeabs/types"
)

// TODO: need to refactor this test, use driven table
func TestParseMsgToMemo(t *testing.T) {
	twapRouter := types.TwapRouter{
		SlippagePercentage: "20",
		WindowSeconds:      10,
	}

	swap := types.Swap{
		OutPutDenom: "khanhyeuchau",
		Slippage:    types.Twap{Twap: twapRouter},
		Receiver:    "123456",
	}

	msgSwap := types.OsmosisSwapMsg{
		OsmosisSwap: swap,
	}

	mockAddress := "cosmos123456789"

	//TODO: need to check assert msg
	_, err := types.ParseMsgToMemo(msgSwap, mockAddress)
	require.NoError(t, err)
}

// TODO: need to refactor this test, use driven table
func TestParseCrossChainSwapMsgToMemo(t *testing.T) {
	outPutDenom := "uosmo"
	contractAddress := "osmo1c3ljch9dfw5kf52nfwpxd2zmj2ese7agnx0p9tenkrryasrle5sqf3ftpg"
	mockReceiver := "feeabs1efd63aw40lxf3n4mhf7dzhjkr453axurwrhrrw"
	chainName := "feeabs"

	execepted_memo_str := `{"wasm":{"contract":"osmo1c3ljch9dfw5kf52nfwpxd2zmj2ese7agnx0p9tenkrryasrle5sqf3ftpg","msg":{"osmosis_swap":{"output_denom":"uosmo","slippage":{"twap":{"slippage_percentage":"20","window_seconds":10}},"receiver":"feeabs/feeabs1efd63aw40lxf3n4mhf7dzhjkr453axurwrhrrw","on_failed_delivery":"do_nothing"}}}}`
	//TODO: need to check assert msg
	memo_str, err := types.BuildCrossChainSwapMemo(outPutDenom, contractAddress, mockReceiver, chainName)

	require.NoError(t, err)
	require.Equal(t, execepted_memo_str, memo_str)
}
