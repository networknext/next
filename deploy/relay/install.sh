#!/bin/bash

bin='relay'
env='relay.env'
svc='relay.service'

bin_dest='/var/relay'
env_dest='/var/relay.env'
svc_dest='/lib/systemd/system/relay.service'

backup_existing() {
	for file in "$@"; do
		if [[ -f "$file" ]]; then
			cp "$file" "$file.backup"
		fi
	done
}

sudo systemctl stop relay

while systemctl is-active --quiet relay; do
	sleep 1
done

backup_existing "$bin_dest" "$env_dest" "$svc_dest"

sudo mv "$bin" "$bin_dest"
sudo mv "$env" "$env_dest"
sudo mv "$svc" "$svc_dest"

sudo systemctl daemon-reload
sudo systemctl start relay

