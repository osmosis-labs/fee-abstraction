package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// IBCPortID is the default port id that profiles module binds to.
	IBCPortID = "feeabs"
)

var ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

// IBCPortKey defines the key to store the port ID in store.
var IBCPortKey = []byte{0x01}

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
