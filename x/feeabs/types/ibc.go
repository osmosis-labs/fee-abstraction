package types

const (
	// IBCPortID is the default port id that profiles module binds to.
	IBCPortID = "feeabs"
)

// IBCPortKey defines the key to store the port ID in store.
var IBCPortKey = []byte{0x01}

type SwapAmountInRoute struct {
	PoolId        uint64
	TokenOutDenom string
}

type OsmosisQueryRequestPacketData struct {
	PoolId  uint64
	TokenIn string
	Routes  []SwapAmountInRoute
}

func NewOsmosisQueryRequestPacketData(poolId uint64, tokenIn string, routes []SwapAmountInRoute) OsmosisQueryRequestPacketData {
	return OsmosisQueryRequestPacketData{
		PoolId:  poolId,
		TokenIn: tokenIn,
		Routes:  routes,
	}
}
