#!/bin/bash

if [[ "$1" == "build" ]]; then
	.vscode/build.sh
	premake5 --file=./premake5.lua gmake
	make
elif [[ "$1" == "run" ]]; then
	export RELAYSPEED="1G"
	export RELAYDEV="1"
	export RELAYMASTER="127.0.0.1"
	export RELAYUPDATEKEY="$RELAY_PUBLIC_KEY" # not correct, just satisfy's the use
	export RELAYADDRESS="127.0.0.1"
	export RELAYPUBLICKEY="$RELAY_PUBLIC_KEY"
	export RELAYPRIVATEKEY="$RELAY_PRIVATE_KEY"
	export RELAYROUTERPUBLICKEY="$RELAY_ROUTER_PUBLIC_KEY" # new to the codebase
	export RELAYBACKENDHOSTNAME="$RELAY_BACKEND_HOSTNAME"

	if [[ "$2" == "one" ]]; then
		export RELAYPORT="20000"
	elif [[ "$2" == "two" ]]; then
		export RELAYPORT="20001"
	fi

	export RELAYNAME="$RELAYADDRESS:$RELAYPORT"

	bin/relay
fi

