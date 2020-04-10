#!/bin/bash

bin='relay'
env='relay.env'
svc='relay.service'

bin_dest="/app/$bin"
env_dest="/app/$env"
svc_dest="/lib/systemd/system/$svc"

backup_existing() {
  time=$(date -u +"%Y%m%d%H%M%S")
	for file in "$@"; do
		if [[ -f "$file" ]]; then
			cp "$file" "$file.$time.backup"
		fi
	done
}

service_is_active() {
  return $(systemctl is-active --quiet relay)
}

service_is_active && sudo systemctl stop relay

while service_is_active; do
	sleep 1
done

if [[ ! -d '/app' ]]; then
  sudo mkdir '/app'
fi

backup_existing "$bin_dest" "$env_dest" "$svc_dest"

sudo mv "$bin" "$bin_dest"
sudo mv "$env" "$env_dest"
sudo mv "$svc" "$svc_dest"

sudo systemctl daemon-reload
sudo systemctl start relay
