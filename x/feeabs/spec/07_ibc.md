# IBCMessages

## `SendQueryIbcDenomTWAP`

A Ibc-token/Native-token TWAP pair is achieved by using the `QueryArithmeticTwapToNowRequest` and `InterchainQueryPacketData`:

```go
type QueryArithmeticTwapToNowRequest struct {
	PoolId     uint64
	BaseAsset  string
	QuoteAsset string
	StartTime  time.Time
}
```

```go
type InterchainQueryPacketData struct {
	Data []byte
	Memo string
}
```

The `QueryArithmeticTwapToNowRequest` will be embedded in the `Data` field of the `InterchainQueryPacketData`

This message will send a query TWAP to the feeabs-contract on counterparty chain (Osmosis) represented by the counterparty Channel End connected to the Channel End with the identifiers `SourcePort` and `SourceChannel`.

The denomination provided for QueryArithmeticTwapToNowRequest should correspond to the same denomination represented on Osmosis.

## `SwapCrossChain`

Feeabs module exchange Ibc token to native token using the `SwapCrossChain` which is `MsgTransfer` with a specific `Memo`:

Swap cross chain shall not be performed if fee abstraction connection to host zone is either FROZEN or OUTDATED. Upon successful twap query, the connection will be set to UPDATED.

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

# Host Chain

Host chain is the swap service provider that fee abstraction uses to swap a token for native fee. Currently, host chain is designed for Osmosis.

<<<<<<< HEAD
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
=======
Fee Abstraction connection to host chain should always be kept alive unless specified FROZEN, as this is crucial to the normal function of fee abstraction.

A host chain config for fee abstraction will contains:

```proto
enum HostChainFeeAbsStatus {
  UPDATED = 0;
  OUTDATED = 1;
  FROZEN = 2;
}

message HostChainFeeAbsConfig {
  // ibc token is allowed to be used as fee token
  string ibc_denom = 1 [ (gogoproto.moretags) = "yaml:\"allowed_token\"" ];
  // token_in in cross_chain swap contract.
  string osmosis_pool_token_denom_in = 2;
  // pool id
  uint64 pool_id = 3;
  // Host chain fee abstraction connection status
  HostChainFeeAbsStatus status = 4;
}
```

1. HostChainFeeAbsStatus
There are four status of fee abstraction connection to host chain:
* UPDATED: the connection is up - to - date.
* OUTDATED: the connection is out of date after failure to ibq query, or fail to cross - chain swap after 5 retries. Should be resumed after 30 mins.
* FROZEN: the connection is frozen, no further actions will be performed.
>>>>>>> d2b5f20 (migrate from frozen to more generic host chain fee abs connection status (#156))
