package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ForwardMetadata struct {
	Receiver string        `json:"receiver,omitempty"`
	Port     string        `json:"port,omitempty"`
	Channel  string        `json:"channel,omitempty"`
	Timeout  time.Duration `json:"timeout,omitempty"`
	Retries  *uint8        `json:"retries,omitempty"`

	// Memo for the cross-chain-swap contract
	Next string `json:"next,omitempty"`
}

func BuildNextMemo(inputToken sdk.Coin, outputDenom string, contractAddress, receiver string) (string, error) {
	swap := Swap{
		InputCoin:   inputToken,
		OutPutDenom: outputDenom,
		Slippage: Twap{
			Twap: TwapRouter{
				SlippagePercentage: "20",
				WindowSeconds:      10,
			},
		},
		Receiver: receiver,
	}

	msgSwap := OsmosisSwapMsg{
		OsmosisSwap: swap,
	}
	return ParseMsgToMemo(msgSwap, contractAddress, receiver)
}
