#!/bin/sh
gaiad tx gov submit-legacy-proposal set-hostzone-config proposals/set_host_zone.json --from validator --gas auto --gas-adjustment 1.5 -y


sleep 5

gaiad tx gov vote  yes --from validator -y