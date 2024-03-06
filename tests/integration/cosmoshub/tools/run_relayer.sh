#!/bin/sh

rm -rf ~/.relayer
rly config init
rly chains add-dir relayers/chains
rly paths add-dir relayers/paths

rly keys restore gaia relayer "bomb jewel cushion behave orphan lava bulk defy evolve spend match grass dress upgrade blast please business stairs learn syrup kick narrow bleak canoe"
rly keys restore osmosis relayer "bomb jewel cushion behave orphan lava bulk defy evolve spend match grass dress upgrade blast please business stairs learn syrup kick narrow bleak canoe"

rly tx link transfer -d -t 10s --client-tp 36h
rly tx link query -d -t 10s --client-tp 36h

rly start
