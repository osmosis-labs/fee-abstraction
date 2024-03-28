package feeabs

import (
	"encoding/base64"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func AttributeValue(txResponse *sdk.TxResponse, eventType, attrKey string) (string, bool) {
	if txResponse == nil {
		return "", false
	}
	for _, event := range txResponse.Events {
		if event.Type != eventType {
			continue
		}
		for _, attr := range event.Attributes {
			if attr.Key == attrKey {
				return attr.Value, true
			}
			key, err := base64.StdEncoding.DecodeString(attr.Key)
			if err != nil {
				continue
			}
			if string(key) == attrKey {
				value, err := base64.StdEncoding.DecodeString(attr.Value)
				if err != nil {
					continue
				}
				return string(value), true
			}
		}
	}
	return "", false
}
