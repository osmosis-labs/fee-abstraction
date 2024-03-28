#!/bin/sh
TX_HASH=$(gaiad tx gov submit-legacy-proposal set-hostzone-config proposals/set_host_zone.json --from validator --gas auto --gas-adjustment 1.5 -y --output json | jq -r '.txhash')
echo "tx hash: $TX_HASH"

PROPOSAL_ID=$(gaiad query tx $TX_HASH --output json | jq -r '.logs[0].events[-1].attributes[-1].value')
echo "proposal id: $PROPOSAL_ID"

gaiad tx gov vote $PROPOSAL_ID yes --from validator -y
