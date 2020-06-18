#!/bin/bash

# $1 = ssh key file
# $2 = port
# $3 = username@address

readonly proj_root="$(pwd)"

readonly dist_dir="$proj_root/dist"

readonly tarfile="relay.tar.gz"

readonly libsodium_file="LATEST.tar.gz"
readonly libsodium_url="https://download.libsodium.org/libsodium/releases/$libsodium_file"

ensure_downloaded() {
	file="$1"
	url="$2"

	if [ ! -f "$file" ]; then
		wget "$url" || return 1
	fi
}

cd "$dist_dir"

# check dependencies afte cd-ing into the dist dir

ensure_downloaded "$libsodium_file" "$libsodium_url" || exit 1

# copy them so they have a generic name

cp "$libsodium_file" 'libsodium.tar.gz'

# tar everything for scp-ing
tar -zcf "$proj_root/dist/$tarfile" 'relay' 'relay.env' 'relay.service' 'install.sh' 'libsodium.tar.gz' || exit 1

scp -i "$1" -P "$2" "$proj_root/dist/$tarfile" "$3:~/$tarfile" || exit 1

# directly run the install script on the relay
ssh -i "$1" -p "$2" "$3" -- "tar -xvf $tarfile && chmod 755 ./install.sh && sudo ./install.sh -i" || exit 1
