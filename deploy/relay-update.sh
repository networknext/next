#!/bin/bash

# $1 = ssh key file
# $2 = username@address

proj_root="$(pwd)"

tarfile="relay.tar.gz"

mkdir -p 'dist/artifact/relay'

cp 'dist/relay' 'dist/artifact/relay/relay'
cp 'deploy/relay/relay.env' 'dist/artifact/relay/relay.env'
cp 'deploy/relay/relay.service' 'dist/artifact/relay/relay.service'
cp 'deploy/relay/install.sh' 'dist/artifact/relay/install.sh'

cd 'dist/artifact/relay'

tar -zcf "$proj_root/dist/$tarfile" 'relay' 'relay.env' 'relay.service' 'install.sh'
scp -i "$1" "$proj_root/dist/$tarfile" "$2:~/$tarfile"
ssh -i "$1" "$2" -- "tar -xvf $tarfile && chmod 755 ./install.sh && sudo ./install.sh -i"
