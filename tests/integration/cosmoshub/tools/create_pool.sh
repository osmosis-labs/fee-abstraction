#!/bin/sh

TX_HASH=$(osmosisd tx gamm create-pool --pool-file pools/pool.json --from relayer --gas auto --gas-adjustment 1.5 --fees 10000uosmo -y --output json | jq -r '.txhash')
echo "tx hash: $TX_HASH"

sleep 5

POOL_ID=$(osmosisd q tx $TX_HASH --output json | jq -r '.logs[0].events[-10].attributes[-1].value')
echo "pool id: $POOL_ID"