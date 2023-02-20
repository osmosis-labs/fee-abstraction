#!/bin/bash
set -e

# always returns true so set -e doesn't exit if it is not running.
killall osmosisd || true
rm -rf $HOME/.osmosisd/

osmosisd config keyring-backend test                                  

# init all three validators
osmosisd init --chain-id=testing validator1 

# create keys for all three validators
osmosisd keys add validator1 --keyring-backend=test 

update_genesis () {    
    cat $HOME/.osmosisd/config/genesis.json | jq "$1" > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
}
echo "lyrics wild earn woman spot rich hen cement trade culture audit amount smoke arm use hollow aerobic correct spirit dolphin tragic all transfer enough" | osmosisd keys add alice --recover --keyring-backend=test 

echo "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry" | osmosisd keys add deployer --recover --keyring-backend=test 

# change staking denom to uosmo
update_genesis '.app_state["staking"]["params"]["bond_denom"]="uosmo"'

# osmo1ekqk6ms4fqf2mfeazju4pcu3jq93lcdsfl0tah
osmosisd add-genesis-account $(osmosisd keys show alice -a --keyring-backend=test ) 100000000000uosmo,100000000000stake,100000000000uatom,2000000uakt 
osmosisd add-genesis-account $(osmosisd keys show deployer -a --keyring-backend=test ) 100000000000uosmo,100000000000stake,100000000000uatom,2000000uakt 

# create validator node with tokens to transfer to the three other nodes
osmosisd add-genesis-account $(osmosisd keys show validator1 -a --keyring-backend=test ) 100000000000uosmo,100000000000stake,100000000000uatom,2000000uakt 
osmosisd gentx validator1 500000000uosmo --keyring-backend=test  --chain-id=testing
osmosisd collect-gentxs 


# update staking genesis
update_genesis '.app_state["staking"]["params"]["unbonding_time"]="240s"'

# update crisis variable to uosmo
update_genesis '.app_state["crisis"]["constant_fee"]["denom"]="uosmo"'

# udpate gov genesis
update_genesis '.app_state["gov"]["voting_params"]["voting_period"]="60s"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="uosmo"'

# update epochs genesis
update_genesis '.app_state["epochs"]["epochs"][1]["duration"]="60s"'

# update poolincentives genesis
update_genesis '.app_state["poolincentives"]["lockable_durations"][0]="120s"'
update_genesis '.app_state["poolincentives"]["lockable_durations"][1]="180s"'
update_genesis '.app_state["poolincentives"]["lockable_durations"][2]="240s"'
update_genesis '.app_state["poolincentives"]["params"]["minted_denom"]="uosmo"'

# update incentives genesis
update_genesis '.app_state["incentives"]["lockable_durations"][0]="1s"'
update_genesis '.app_state["incentives"]["lockable_durations"][1]="120s"'
update_genesis '.app_state["incentives"]["lockable_durations"][2]="180s"'
update_genesis '.app_state["incentives"]["lockable_durations"][3]="240s"'
update_genesis '.app_state["incentives"]["params"]["distr_epoch_identifier"]="day"'

# update mint genesis
update_genesis '.app_state["mint"]["params"]["mint_denom"]="uosmo"'
update_genesis '.app_state["mint"]["params"]["epoch_identifier"]="day"'

# update gamm genesis
update_genesis '.app_state["gamm"]["params"]["pool_creation_fee"][0]["denom"]="uosmo"'

# update interchainquery genesis
update_genesis '.app_state["interchainquery"]["params"]["allow_queries"][0]="/cosmos.bank.v1beta1.Query/AllBalances"'


# port key (validator1 uses default ports)
# validator1 1317, 9090, 9091, 26658, 26657, 26656, 6060
# validator2 1316, 9088, 9089, 26655, 26654, 26653, 6061
# validator3 1315, 9086, 9087, 26652, 26651, 26650, 6062

# change config.toml values
VALIDATOR1_CONFIG=$HOME/.osmosisd/config/config.toml

# validator1
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR1_CONFIG

# start all three validators
osmosisd start 

echo "1 Validators are up and running!"