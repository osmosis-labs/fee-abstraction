#!/bin/sh

osmosisd tx gamm create-pool --pool-file pools/pool.json --from relayer --gas auto --gas-adjustment 1.5 --fees 10000uosmo -y