package types

const (
	// Module name store the name of the module
	ModuleName = "feeabs"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the feeabs module
	RouterKey = ModuleName

	// Contract: Coin denoms cannot contain this character
	KeySeparator = "|"
)

var (
	OsmosisExchangeRate = []byte{0x01} // Key for the exchange rate of osmosis (to native token)
	KeyChannelID        = []byte{0x02} // Key for IBC channel to osmosis
)

// GetOsmosisExchangeRateKey return the key for set/getting the exchange rate of osmosis (to native token)
func GetOsmosisExchangeRateKey() (key []byte) {
	return OsmosisExchangeRate
}
