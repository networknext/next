#!/bin/bash

# $1 = ssh key file
# $2 = port
# $3 = username@address

proj_root="$(pwd)"

tarfile="relay.tar.gz"

cd 'dist'

tar -zcf "$proj_root/dist/$tarfile" 'relay' 'relay.env' 'relay.service' 'install.sh' || exit 1
scp -i "$1" -P "$2" "$proj_root/dist/$tarfile" "$3:~/$tarfile" || exit 1
ssh -i "$1" -p "$2" "$3" -- "tar -xvf $tarfile && chmod 755 ./install.sh && sudo ./install.sh -i" || exit 1
