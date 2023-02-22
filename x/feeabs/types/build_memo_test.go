package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/notional-labs/feeabstraction/v1/x/feeabs/types"
)

// TODO: need to refactor this test, use driven table
func TestParseMsgToMemo(t *testing.T) {
	twapRouter := types.TwapRouter{
		SlippagePercentage: "20",
		WindowSeconds:      10,
	}

	swap := types.Swap{
		InputCoin:   sdk.NewCoin("khanhyeungan", sdk.NewInt(123)),
		OutPutDenom: "khanhyeuchau",
		Slippage:    types.Twap{Twap: twapRouter},
		Receiver:    "123456",
	}

	msgSwap := types.OsmosisSwapMsg{
		OsmosisSwap: swap,
	}

	mockAddress := "cosmos123456789"
	mockReceiver := "cosmos123456789"

	//TODO: need to check assert msg
	_, err := types.ParseMsgToMemo(msgSwap, mockAddress, mockReceiver)
	require.NoError(t, err)
}

// TODO: need to refactor this test, use driven table
func TestParseCrossChainSwapMsgToMemo(t *testing.T) {
	inputToken := sdk.NewCoin("stake", sdk.NewInt(123))
	outPutDenom := "uosmo"
	contractAddress := "osmo1c3ljch9dfw5kf52nfwpxd2zmj2ese7agnx0p9tenkrryasrle5sqf3ftpg"
	mockReceiver := "osmo1cd4nn8yzdrrsfqsmmvaafq8r03xn38qgqt8fzh"

	execepted_memo_str := `{"wasm":{"contract":"osmo1c3ljch9dfw5kf52nfwpxd2zmj2ese7agnx0p9tenkrryasrle5sqf3ftpg","msg":{"osmosis_swap":{"input_coin":{"denom":"stake","amount":"123"},"output_denom":"uosmo","slippage":{"twap":{"slippage_percentage":"20","window_seconds":10}},"receiver":"osmo1cd4nn8yzdrrsfqsmmvaafq8r03xn38qgqt8fzh"}},"receiver":"osmo1cd4nn8yzdrrsfqsmmvaafq8r03xn38qgqt8fzh"}}`
	//TODO: need to check assert msg
	memo_str, err := types.BuildCrossChainSwapMemo(inputToken, outPutDenom, contractAddress, mockReceiver)

	require.NoError(t, err)
	require.Equal(t, memo_str, execepted_memo_str)
}

// TODO: need to refactor this test, use driven table
func TestParsePacketMiddlewareMemoToMemo(t *testing.T) {
	inputToken := sdk.NewCoin("stake", sdk.NewInt(123))
	outPutDenom := "uosmo"
	contractAddress := "osmo1c3ljch9dfw5kf52nfwpxd2zmj2ese7agnx0p9tenkrryasrle5sqf3ftpg"
	execepted_memo_str := `{"receiver":"cosmos123","port":"transfer","channel":"channel-56","timeout":1800000,"retries":8,"next":"{\"wasm\":{\"contract\":\"\",\"msg\":{\"osmosis_swap\":{\"input_coin\":{\"denom\":\"stake\",\"amount\":\"123\"},\"output_denom\":\"uosmo\",\"slippage\":{\"twap\":{\"slippage_percentage\":\"20\",\"window_seconds\":10}},\"receiver\":\"osmo1c3ljch9dfw5kf52nfwpxd2zmj2ese7agnx0p9tenkrryasrle5sqf3ftpg\"}},\"receiver\":\"osmo1c3ljch9dfw5kf52nfwpxd2zmj2ese7agnx0p9tenkrryasrle5sqf3ftpg\"}}"}`

	config := types.HostChainFeeAbsConfig{
		IbcDenom:                   "ibc/123",
		OsmosisPoolTokenDenomIn:    "ibc/456",
		MiddlewareAddress:          "cosmos123",
		HostZoneIbcTransferChannel: "channel-56",
	}

	// TODO: need to check assert msg
	memo_str, err := types.BuildPacketMiddlewareMemo(inputToken, outPutDenom, contractAddress, config)

	require.NoError(t, err)
	require.Equal(t, memo_str, execepted_memo_str)
}
