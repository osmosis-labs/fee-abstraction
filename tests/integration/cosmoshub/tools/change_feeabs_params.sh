#!/bin/sh

gaiad tx gov submit-legacy-proposal param-change proposals/params.json --from validator --gas auto --gas-adjustment 1.5 -y



sleep 5

gaiad tx gov vote 1 yes --from validator -y