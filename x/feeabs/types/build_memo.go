package types

import (
	"encoding/json"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
)

type OsmosisSpecialMemo struct {
	Wasm map[string]interface{} `json:"wasm"`
}

type OsmosisSwapMsg struct {
	OsmosisSwap Swap `json:"osmosis_swap"`
}
type Swap struct {
	InputCoin        sdk.Coin `json:"input_coin"`
	OutPutDenom      string   `json:"output_denom"`
	Slippage         Twap     `json:"slippage"`
	Receiver         string   `json:"receiver"`
	OnFailedDelivery string   `json:"on_failed_delivery"`
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

// TODO: write test for this.
func BuildPacketMiddlewareMemo(inputToken sdk.Coin, outputDenom string, receiver string, hostChainConfig HostChainFeeAbsConfig) (string, error) {
	// TODO: this should be chain params.
	timeOut := time.Duration(1800000)
	retries := uint8(8)
	nextMemo, err := BuildCrossChainSwapMemo(inputToken, outputDenom, hostChainConfig.CrosschainSwapAddress, receiver)
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
	return BuildForwardMetaMemo(metadata)
}

type ForwardMetadata struct {
	Receiver string        `json:"receiver,omitempty"`
	Port     string        `json:"port,omitempty"`
	Channel  string        `json:"channel,omitempty"`
	Timeout  time.Duration `json:"timeout,omitempty"`
	Retries  *uint8        `json:"retries,omitempty"`

	// Memo for the cross-chain-swap contract
	Next string `json:"next,omitempty"`
}

// TODO: write test for this
// BuildNextMemo create memo for IBC hook, this execute `CrossChainSwap contract`
func BuildCrossChainSwapMemo(inputToken sdk.Coin, outputDenom string, contractAddress, receiver string) (string, error) {
	swap := Swap{
		InputCoin:   inputToken,
		OutPutDenom: outputDenom,
		Slippage: Twap{
			Twap: TwapRouter{
				SlippagePercentage: "20",
				WindowSeconds:      10,
			},
		},
		Receiver:         receiver,
		OnFailedDelivery: "do_nothing",
	}

	msgSwap := OsmosisSwapMsg{
		OsmosisSwap: swap,
	}
	nextMemo, err := ParseMsgToMemo(msgSwap, contractAddress, receiver)
	if err != nil {
		return "", err
	}

	return nextMemo, nil
}

// TODO: write test for this
func BuildForwardMetaMemo(forwardMetadata ForwardMetadata) (string, error) {
	memo_marshalled, err := json.Marshal(&forwardMetadata)
	if err != nil {
		return "", err
	}
	return string(memo_marshalled), nil
}
