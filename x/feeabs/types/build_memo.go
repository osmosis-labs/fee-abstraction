package types

import (
	"encoding/json"
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type OsmosisSpecialMemo struct {
	Wasm map[string]interface{} `json:"wasm"`
}

type OsmosisSwapMsg struct {
	OsmosisSwap Swap `json:"osmosis_swap"`
}
type Swap struct {
	OutPutDenom      string `json:"output_denom"`
	Slippage         Twap   `json:"slippage"`
	Receiver         string `json:"receiver"`
	OnFailedDelivery string `json:"on_failed_delivery"`
}

type Twap struct {
	Twap TwapRouter `json:"twap"`
}

type TwapRouter struct {
	SlippagePercentage string `json:"slippage_percentage"`
	WindowSeconds      uint64 `json:"window_seconds"`
}

type PacketMetadata struct {
	Forward *ForwardMetadata `json:"forward"`
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

func NewOsmosisSwapMsg(inputCoin sdk.Coin, outputDenom string, slippagePercentage string, windowSeconds uint64, receiver string) OsmosisSwapMsg {
	swap := Swap{
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
func ParseMsgToMemo(msg OsmosisSwapMsg, contractAddr string, receiverAddress string, chainName string) (string, error) {
	// TODO: need to validate the msg && contract address
	receiver := fmt.Sprintf("%s/%s", chainName, receiverAddress)
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

// TODO: write test for this
// BuildNextMemo create memo for IBC hook, this execute `CrossChainSwap contract`
func BuildCrossChainSwapMemo(outputDenom string, contractAddress, receiver string, chainName string) (string, error) {
	swap := Swap{
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
	nextMemo, err := ParseMsgToMemo(msgSwap, contractAddress, receiver, chainName)
	if err != nil {
		return "", err
	}

	return nextMemo, nil
}
