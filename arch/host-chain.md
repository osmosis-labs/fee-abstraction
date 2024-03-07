# Host Chain

Host chain is the swap service provider that fee abstraction uses to swap a token for native fee. Currently, host chain is designed for Osmosis.

Fee Abstraction connection to host chain should always be kept alive unless specified FROZEN, as this is crucial to the normal function of fee abstraction.

A host chain config for fee abstraction will contains:

```proto
enum HostChainFeeAbsStatus {
  UNSPECIFIED = 0;
  UPDATED = 1;
  OUTDATED = 2;
  FROZEN = 3;
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
* UNSPECIFIED: the connection is unspecified, and will be determined by the chain.
* UPDATED: the connection is up - to - date.
* OUTDATED: the connection is out of date after failure to ibq query, or fail to cross - chain swap after 5 retries.
* FROZEN: the connection is frozen, no further actions will be performed.

