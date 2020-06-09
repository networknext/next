#!/bin/bash

# $1 = ssh key file
# $2 = username@address

proj_root="$(pwd)"

tarfile="relay.tar.gz"

cd 'dist'

tar -zcf "$proj_root/dist/$tarfile" 'relay' 'relay.env' 'relay.service' 'install.sh'
scp -i "$1" "$proj_root/dist/$tarfile" "$2:~/$tarfile"
ssh -i "$1" "$2" -- "tar -xvf $tarfile && chmod 755 ./install.sh && sudo ./install.sh -i"
