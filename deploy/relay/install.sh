#!/bin/bash

bin='relay'
env='relay.env'
svc='relay.service'
app='/app'

bin_dest="$app/$bin"
env_dest="$app/$env"
svc_dest="/lib/systemd/system/$svc"

print_help() {
	printf "Relay installer script\n\n"
	printf "i)\tinstall a new relay\n"
	printf "r)\trevert to the last installed version\n"
}

backup_existing() {
	time=$(date -u +"%Y%m%d%H%M%S")
	for file in "$@"; do
		if [[ -f "$file" ]]; then
      bname="$( basename $file )"
			cp "$file" "$app/$bname.$time.backup"
		fi
	done
}

check_if_running() {
	err_msg="$1"
	if systemctl is-active --quiet relay; then
		echo "$err_msg"
		exit 1
	fi
}

install_relay() {
	check_if_running 'error, please disable relay before installing'

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
}

revert_relay() {
	check_if_running 'error, please disable relay before reverting'

  cd '/app'

  relays=()
  while IFS= read -d $'\0' -r relay; do
    relays=("${relays[@]}" "$relay")
  done < <( find . -regextype posix-extended -regex '.*/relay\.[0-9]+\.backup' -print0 )

  if ! (( ${#relays[@]} )); then
    echo 'no relay to revert to'
    exit 1
  fi

  # get the most recent relay binary
  relay="${relays[-1]}"

  echo "reverting to relay '$relay'"

  # match the timestamps
  env_file="${relay/relay/relay.env}"
  svc_file="${relay/relay/relay.service}"

  echo "looking for environment file '$env_file'"

  if [ ! -f "$env_file" ]; then
    echo 'no environment file to revert to'
    exit 1
  fi

  echo "looking for environment file '$svc_file'"

  if [ ! -f "$svc_file" ]; then
    echo 'no service file to revert to'
    exit 1
  fi

  # if a matching relay, environment file, and service file all exist then replace with them
  mv "$relay" "$bin_dest"
  mv "$env_file" "$env_dest"
  mv "$svc_file" "$svc_dest"

	sudo systemctl daemon-reload
	sudo systemctl enable relay
	sudo systemctl start relay
}

cmd=''

while getopts 'irh' flag; do
	case ${flag} in
		i) cmd='i' ;;
		r) cmd='r' ;;
		h)
			print_help
			exit 1
			;;
		*)
			print_help
			exit 1
			;;
	esac
done

if [ "$cmd" = 'i' ]; then
	install_relay
elif [ "$cmd" = 'r' ]; then
	revert_relay
fi
