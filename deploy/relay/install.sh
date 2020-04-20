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

if systemctl is-active --quiet relay; then
  echo "Error: Please disable the relay before updating"
  exit 1
fi

if [[ ! -d '/app' ]]; then
  sudo mkdir '/app'
fi

backup_existing "$bin_dest" "$env_dest" "$svc_dest"

sudo mv "$bin" "$bin_dest"
sudo mv "$env" "$env_dest"
sudo mv "$svc" "$svc_dest"

sudo systemctl daemon-reload
sudo systemctl enable relay
sudo systemctl start relay
