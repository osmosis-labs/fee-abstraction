package types

const (
	// IBCPortID is the default port id that profiles module binds to.
	IBCPortID = "feeabs"
)

// IBCPortKey defines the key to store the port ID in store.
var IBCPortKey = []byte{0x01}
