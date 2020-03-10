#!/bin/bash

if [[ "$1" == "build" ]]; then
	.vscode/build.sh # will typically throw an error because premake5 won't sense the lua file even though it is named right
	premake5 --file=./premake5.lua gmake # hence this, except explicitly specifying the lua file
	make # finally make the thing
elif [[ "$1" == "run" ]]; then
	export RELAYSPEED="1G"
	export RELAYDEV="1" # 1 || 0 for enabling/disabling dev mode
	export RELAYMASTER="127.0.0.1" # ip addr of the backend, unused with http. pretty sure this is an accurate assumption but I could be wrong
	export RELAYUPDATEKEY="$RELAY_PUBLIC_KEY" # not correct, just satisfy's the internal use, not used with http
	export RELAYADDRESS="127.0.0.1" # must not have port
	export RELAYPUBLICKEY="$RELAY_PUBLIC_KEY" # actual relay public key
	export RELAYPRIVATEKEY="$RELAY_PRIVATE_KEY" # actual relay private key
	export RELAYROUTERPUBLICKEY="$RELAY_ROUTER_PUBLIC_KEY" # new to the codebase, made the var similar to the others for the sake of consitency, original router key is hardcoded at top of relay_internal.cpp
	export RELAYBACKENDHOSTNAME="$RELAY_BACKEND_HOSTNAME" # ditto, except 

	# quick & lazy way to just spawn 2 relays on different ports
	if [[ "$2" == "one" ]]; then
		export RELAYPORT="20000"
	elif [[ "$2" == "two" ]]; then
		export RELAYPORT="20001"
	fi

	export RELAYNAME="$RELAYADDRESS:$RELAYPORT" # originally these were names, but now the id is the address:port hash instead of the name hash

	bin/relay
fi

