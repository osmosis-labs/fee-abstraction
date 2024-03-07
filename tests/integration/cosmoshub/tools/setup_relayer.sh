#!/bin/sh

rm -rf ~/.relayer
rly config init
rly chains add-dir relayers/chains

rly keys restore gaia relayer "bomb jewel cushion behave orphan lava bulk defy evolve spend match grass dress upgrade blast please business stairs learn syrup kick narrow bleak canoe"
rly keys restore osmosis relayer "bomb jewel cushion behave orphan lava bulk defy evolve spend match grass dress upgrade blast please business stairs learn syrup kick narrow bleak canoe"
