#!/bin/sh
gaiad tx gov submit-legacy-proposal add-hostzone-config proposals/add_host_zone.json --from validator --gas auto --gas-adjustment 1.5 -y


sleep 5

gaiad tx gov vote 2 yes --from validator -y