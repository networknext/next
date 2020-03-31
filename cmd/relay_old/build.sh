#!/bin/bash

set -e
set -x

# you also need these packages installed:
# sudo apt install build-essentials gdb libsparsehash-dev

if [[ "$1" == "build" ]]; then
	if [ "$(which premake5)" == "" ]; then
		cd ~
		rm -Rf ~/premake-* || true
		rm -f 'premake5' || true
		curl -sSL -o 'premake-linux.tar.gz https://github.com/premake/premake-core/releases/download/v5.0.0-alpha14/premake-5.0.0-alpha14-linux.tar.gz'
		tar -xvf 'premake-linux.tar.gz'
		rm 'premake-linux.tar.gz'
		chmod a+x premake5
		sudo mv premake5 '/usr/local/bin/premake5'
	fi

	premake5 --file=./premake5.lua gmake && make -j32
	make # finally make the thing
elif [[ "$1" == "run" ]]; then
	export RELAYSPEED="1G"
	export RELAYDEV="1" # 1 for enabling dev mode || !1 for disabling dev mode, dev mode doesn't kill your system
	export RELAYMASTER="127.0.0.1" # ip addr of the backend, unused with http since that requires http + this and I was lazy
	export RELAYUPDATEKEY="$RELAY_PUBLIC_KEY" # not correct, just satisfy's the base64 decryption, not used with http, comes from firestore
	export RELAYPUBLICKEY="$RELAY_PUBLIC_KEY" # actual relay public key
	export RELAYPRIVATEKEY="$RELAY_PRIVATE_KEY" # actual relay private key
	export RELAYROUTERPUBLICKEY="$RELAY_ROUTER_PUBLIC_KEY" # new to the codebase, made the var similar to the others for the sake of consitency, original router key is hardcoded at top of relay_internal.cpp
	export RELAYBACKENDHOSTNAME="$RELAY_BACKEND_HOSTNAME" # ditto, except nothing is hard coded, just first two comma sections

	# quick & lazy way to just spawn 2 relays on different ports
	if [[ "$2" == "one" ]]; then
		export RELAYPORT="20000" # after init this is set on the each env's relay address port
	elif [[ "$2" == "two" ]]; then
		export RELAYPORT="20001"
	fi

	export RELAYADDRESS="127.0.0.1:$RELAYPORT" # must have port

	export RELAYNAME="$RELAYADDRESS" # originally these were names, but now the id is the address:port hash instead of the name hash

	bin/relay
fi

