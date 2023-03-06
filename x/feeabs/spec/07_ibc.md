# IBCMessages

## `SendQueryIbcDenomTWAP`

A Ibc-token/Native-token TWAP pair is achieved by using the `QueryArithmeticTwapToNowRequest` and `InterchainQueryRequest`:

```go
type QueryArithmeticTwapToNowRequest struct {
	PoolId     uint64
	BaseAsset  string
	QuoteAsset string
	StartTime  time.Time
}
```

```go
type InterchainQueryRequest struct {
	Data []byte
	Path string
}
```

The `QueryArithmeticTwapToNowRequest` will be embedded in the `Data` field of the `InterchainQueryRequest`, the `Path` field should match with the Osmosis `ArithmeticTwapToNow` GRPC query, which is `"/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow"`

This message will send a query TWAP to the feeabs-contract on counterparty chain (Osmosis) represented by the counterparty Channel End connected to the Channel End with the identifiers `SourcePort` and `SourceChannel`.

The denomination provided for QueryArithmeticTwapToNowRequest should correspond to the same denomination represented on Osmosis.

## `SwapCrossChain`

Feeabs module exchange Ibc token to native token using the `SwapCrossChain`:

```go
type MsgTransfer struct {
	SourcePort string
	SourceChannel string
	Token types.Coin
	Sender string
	Receiver string
	TimeoutHeight types1.Height
	TimeoutTimestamp uint64
	Memo string
}
```

This message is expected to fail if:

- `SourcePort` is invalid (see [24-host naming requirements](https://github.com/cosmos/ibc/blob/master/spec/core/ics-024-host-requirements/README.md#paths-identifiers-separators).
- `SourceChannel` is invalid (see [24-host naming requirements](https://github.com/cosmos/ibc/blob/master/spec/core/ics-024-host-requirements/README.md#paths-identifiers-separators)).
- `Token` is invalid (denom is invalid or amount is negative)
  - `Token.Amount` is not positive.
  - `Token.Denom` is not a valid IBC denomination as per [ADR 001 - Coin Source Tracing](../../../docs/architecture/adr-001-coin-source-tracing.md).
- `Sender` is empty.
- `Receiver` is empty.
- `TimeoutHeight` and `TimeoutTimestamp` are both zero.

Feeabs module will send an ibc transfer message with a sepecific data in `Memo` field. This `Memo` field data will be used in Ibc transfer middleware on counterparty chain to swap the amount of ibc token to native token on Osmosis.

There will be 2 separate case that the counterparty chain is Osmosis or not, we will have 2 correspond `Memo`.

These 2 case are defined in the `IsOsmosis` field in `HostChainFeeAbsConfig`

```go
type HostChainFeeAbsConfig struct {
	IbcDenom string
	OsmosisPoolTokenDenomIn string
	MiddlewareAddress string
	IbcTransferChannel string
	HostZoneIbcTransferChannel string
	CrosschainSwapAddress string
	PoolId uint64
	IsOsmosis bool
	Frozen bool
	OsmosisQueryChannel string
}
```
Note: These 2 Ibc message only open for testing version. In the product version, user can't manual send these 2 message instead, feeabs module will automatic send every epoch to update the TWAP and swap ibc-token to native-token.