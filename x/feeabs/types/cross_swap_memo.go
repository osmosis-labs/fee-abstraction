package types

import (
	"encoding/json"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
)

const (
	// IBCPortID is the default port id that profiles module binds to.
	IBCPortID = "feeabs"
)

type SpotPrice struct {
	SpotPrice string `json:"spot_price"`
}

var ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

// IBCPortKey defines the key to store the port ID in store.
var (
	IBCPortKey        = []byte{0x01}
	FeePoolAddressKey = []byte{0x02}
)

// NewOsmosisQueryRequestPacketData create new packet for ibc.
func NewOsmosisQueryRequestPacketData(poolId uint64, baseDenom string, quoteDenom string) OsmosisQuerySpotPriceRequestPacketData {
	return OsmosisQuerySpotPriceRequestPacketData{
		PoolId:          poolId,
		BaseAssetDenom:  baseDenom,
		QuoteAssetDenom: quoteDenom,
	}
}

// GetBytes is a helper for serializing.
func (p OsmosisQuerySpotPriceRequestPacketData) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&p))
}

// TODO: Those types should be putted in types package
// `{
// 	"wasm": {
// 	  "contract": "CROSSCHAIN_SWAPS_ADDRESS",
// 	  "msg": {
// 		"osmosis_swap": {
// 		  "input_coin": {
// 			"denom": "$DENOM",
// 			"amount": "100"
// 		  },
// 		  "output_denom": "uosmo",
// 		  "slippage": {
// 			"twap": {
// 			  "slippage_percentage": "20",
// 			  "window_seconds": 10
// 			}
// 		  },
// 		  "receiver": "$VALIDATOR"
// 		}
// 	  }
// 	}
//   }
//   `
type OsmosisSpecialMemo struct {
	Wasm map[string]interface{} `json:"wasm"`
}

type OsmosisSwapMsg struct {
	OsmosisSwap Swap `json:"osmosis_swap"`
}
type Swap struct {
	InputCoin   sdk.Coin `json:"input_coin"`
	OutPutDenom string   `json:"output_denom"`
	Slippage    Twap     `json:"slippage"`
	Receiver    string   `json:"receiver"`
}

type Twap struct {
	Twap TwapRouter `json:"twap"`
}

type TwapRouter struct {
	SlippagePercentage string `json:"slippage_percentage"`
	WindowSeconds      uint64 `json:"window_seconds"`
}

func NewOsmosisSwapMsg(inputCoin sdk.Coin, outputDenom string, slippagePercentage string, windowSeconds uint64, receiver string) OsmosisSwapMsg {
	swap := Swap{
		InputCoin:   inputCoin,
		OutPutDenom: outputDenom,
		Slippage: Twap{
			Twap: TwapRouter{SlippagePercentage: slippagePercentage,
				WindowSeconds: windowSeconds,
			}},
		Receiver: receiver,
	}

	return OsmosisSwapMsg{
		OsmosisSwap: swap,
	}
}

// ParseMsgToMemo build a memo from msg, contractAddr, compatible with ValidateAndParseMemo in https://github.com/osmosis-labs/osmosis/blob/nicolas/crosschain-swaps-new/x/ibc-hooks/wasm_hook.go
func ParseMsgToMemo(msg OsmosisSwapMsg, contractAddr string, receiver string) (string, error) {
	// TODO: need to validate the msg && contract address
	memo := OsmosisSpecialMemo{
		Wasm: make(map[string]interface{}),
	}

	memo.Wasm["contract"] = contractAddr
	memo.Wasm["msg"] = msg
	memo.Wasm["receiver"] = receiver

	memo_marshalled, err := json.Marshal(&memo)
	if err != nil {
		return "", err
	}
	return string(memo_marshalled), nil
}

func BuildPacketMiddlewareMemo(inputToken sdk.Coin, outputDenom string, receiver string, hostChainConfig HostChainFeeAbsConfig) (string, error) {
	// TODO: this should be chain params.
	timeOut := time.Duration(1800000)
	retries := uint8(8)
	nextMemo, err := BuildNextMemo(inputToken, outputDenom, hostChainConfig.CrosschainSwapAddress, receiver)
	if err != nil {
		return "", nil
	}
	metadata := ForwardMetadata{
		Receiver: hostChainConfig.MiddlewareAddress,
		Port:     transfertypes.PortID,
		Channel:  hostChainConfig.HostZoneIbcTransferChannel,
		Timeout:  timeOut,
		Retries:  &retries,
		Next:     nextMemo,
	}

	// TODO: need to validate the msg && contract address.
	memo_marshalled, err := json.Marshal(&metadata)
	if err != nil {
		return "", err
	}
	return string(memo_marshalled), nil
}
