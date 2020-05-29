#!/bin/bash

# $1 = ssh key file
# $2 = username@address

# Once we start moving into prod and have updated all of the relays at least once,
# we can remove the two outer checks and just use the innermost check

# Check for update key in relay.service.backup file (for dev relays)
updateKey=$(ssh -i "$1" "$2" -- "grep UPDATEKEY_0 /lib/systemd/system/relay.service.backup | sed 's:^Environment=\"RELAYUPDATEKEY_0=::; s:\"$::'")

if [ -z $updateKey ]; then
    # Check for update key in relay.service file (for prod relays first time)
    updateKey=$(ssh -i "$1" "$2" -- "grep UPDATEKEY_0 /lib/systemd/system/relay.service | sed 's:^Environment=\"RELAYUPDATEKEY_0=::; s:\"$::'")

    if [ -z $updateKey ]; then
        # Check for update key in relay.env file (for prod relays after first time)
        updateKey=$(ssh -i "$1" "$2" -- "grep RELAY_V3_UPDATE_KEY /app/relay.env | sed 's:^RELAY_V3_UPDATE_KEY=::'")

        if [ -z $updateKey ]; then
            echo "could not find update key in relay.service.backup or relay.service"
            exit 1
        fi
    fi
fi

echo $updateKey
