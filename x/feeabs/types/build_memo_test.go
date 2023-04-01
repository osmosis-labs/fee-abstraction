package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/notional-labs/feeabstraction/v2/x/feeabs/types"
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

	execepted_memo_str := `{"wasm":{"contract":"osmo1c3ljch9dfw5kf52nfwpxd2zmj2ese7agnx0p9tenkrryasrle5sqf3ftpg","msg":{"osmosis_swap":{"input_coin":{"denom":"stake","amount":"123"},"output_denom":"uosmo","slippage":{"twap":{"slippage_percentage":"20","window_seconds":10}},"receiver":"osmo1cd4nn8yzdrrsfqsmmvaafq8r03xn38qgqt8fzh","on_failed_delivery":"do_nothing"}},"receiver":"osmo1cd4nn8yzdrrsfqsmmvaafq8r03xn38qgqt8fzh"}}`
	//TODO: need to check assert msg
	memo_str, err := types.BuildCrossChainSwapMemo(inputToken, outPutDenom, contractAddress, mockReceiver)

	require.NoError(t, err)
	require.Equal(t, execepted_memo_str, memo_str)
}

// TODO: need to refactor this test, use driven table
func TestParsePacketMiddlewareMemoToMemo(t *testing.T) {
	inputToken := sdk.NewCoin("stake", sdk.NewInt(123))
	outputDenom := "uosmo"
	contractAddress := "osmo1c3ljch9dfw5kf52nfwpxd2zmj2ese7agnx0p9tenkrryasrle5sqf3ftpg"
	mockReceiver := "osmo1cd4nn8yzdrrsfqsmmvaafq8r03xn38qgqt8fzh"
	execepted_memo_str := `{"forward":{"receiver":"osmo1c3ljch9dfw5kf52nfwpxd2zmj2ese7agnx0p9tenkrryasrle5sqf3ftpg","port":"transfer","channel":"channel-56","timeout":600000000000,"retries":0,"next":"{\"wasm\":{\"contract\":\"osmo1c3ljch9dfw5kf52nfwpxd2zmj2ese7agnx0p9tenkrryasrle5sqf3ftpg\",\"msg\":{\"osmosis_swap\":{\"input_coin\":{\"denom\":\"stake\",\"amount\":\"123\"},\"output_denom\":\"uosmo\",\"slippage\":{\"twap\":{\"slippage_percentage\":\"20\",\"window_seconds\":10}},\"receiver\":\"osmo1cd4nn8yzdrrsfqsmmvaafq8r03xn38qgqt8fzh\",\"on_failed_delivery\":\"do_nothing\"}},\"receiver\":\"osmo1cd4nn8yzdrrsfqsmmvaafq8r03xn38qgqt8fzh\"}}"}}`

	config := types.HostChainFeeAbsConfig{
		IbcDenom:                   "ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E",
		OsmosisPoolTokenDenomIn:    "ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E",
		MiddlewareAddress:          "cosmos1alc8mjana7ssgeyffvlfza08gu6rtav8rmj6nv",
		IbcTransferChannel:         "channel-2",
		HostZoneIbcTransferChannel: "channel-56",
		CrosschainSwapAddress:      contractAddress,
		PoolId:                     1,
		IsOsmosis:                  false,
		Frozen:                     false,
		OsmosisQueryChannel:        "channel-1",
	}

	// TODO: need to check assert msg
	memo_str, err := types.BuildPacketMiddlewareMemo(inputToken, outputDenom, mockReceiver, config)

	require.NoError(t, err)
	require.Equal(t, execepted_memo_str, memo_str)
}
