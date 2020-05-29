#!/bin/bash

# $1 = ssh key file
# $2 = username@address

updateKey=$(ssh -i "$1" "$2" -- "grep UPDATEKEY_0 /lib/systemd/system/relay.service.backup | sed 's:^Environment=\"RELAYUPDATEKEY_0=::; s:\"$::'")

if [ -z $updateKey ]; then
    updateKey=$(ssh -i "$1" "$2" -- "grep UPDATEKEY_0 /lib/systemd/system/relay.service | sed 's:^Environment=\"RELAYUPDATEKEY_0=::; s:\"$::'")

    if [ -z $updateKey ]; then
        echo "could not find update key in relay.service.backup or relay.service"
        exit 1
    fi
fi

echo $updateKey
