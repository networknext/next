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
    # file is abs path, so reduce to just the filename
    bname="$(basename $file)"
		if [[ -f "$file" ]]; then
			cp "$file" "$app/$bname.$time.backup"
    else
      echo "no '$bname' file found to backup"
		fi
	done
}

check_if_running() {
  # $1 = error message to print before exit
	if systemctl is-active --quiet relay; then
		echo "$1"
		exit 1
	fi
}

install_relay() {
	check_if_running 'error, please disable relay before installing'

	if [[ ! -d '/app' ]]; then
		sudo mkdir '/app' || return 1
	fi

	backup_existing "$bin_dest" "$env_dest" "$svc_dest"

	echo 'installing relay...'
	sudo mv "$bin" "$bin_dest"
	sudo mv "$env" "$env_dest"
	sudo mv "$svc" "$svc_dest"

	sudo systemctl daemon-reload || return 1

	sudo systemctl enable relay || return 1
	sudo systemctl start relay || return 1
	echo 'done'
}

revert_relay() {
	check_if_running 'error, please disable relay before reverting'

  cd '/app'

  # find all backup relays in the /app directory and store them into an array
  relays=()
  while IFS= read -d $'\0' -r relay; do
    relays=("${relays[@]}" "$relay")
  done < <( find . -regextype posix-extended -regex '.*/relay\.[0-9]+\.backup' -print0 )

  # if the length of the array is 0
  if ! (( ${#relays[@]} )); then
    echo 'no relay to revert to'
    return 1
  fi

  # get the most recent relay binary using negative indexing
  relay="${relays[-1]}"

  echo "reverting to relay '$relay'"

  # match the timestamps, replaces relay.xyz.backup with relay.env.xyz.backup, same for svc file
  env_file="${relay/relay/relay.env}"
  svc_file="${relay/relay/relay.service}"

  echo "looking for matching environment file '$env_file'"

  # the relays on the prod systems now don't have env files, so if reverting to them
  # it will not exist and thus the script should not exit because thats a valid case
  if [ -f "$env_file" ]; then
    mv "$env_file" "$env_dest"
  else
    echo 'no environment file to revert to, skipping'
  fi

  echo "looking for matching service file '$svc_file'"

  # however a service file is present for all versions so exit if this is not found
  if [ ! -f "$svc_file" ]; then
    echo 'no service file to revert to'
    return 1
  fi

  echo 'found matching service file'

  mv "$relay" "$bin_dest"
  mv "$svc_file" "$svc_dest"

  # enable and start the relay service
	sudo systemctl daemon-reload || return 1
	sudo systemctl enable relay || return 1
	sudo systemctl start relay || return 1
}

cmd=''

while getopts 'irh' flag; do
	case ${flag} in
		i) cmd='i' ;;
		r) cmd='r' ;;
		h) print_help
       exit 1
       ;;
		*) print_help
       exit 1
       ;;
	esac
done

if [ "$cmd" = 'i' ]; then
	install_relay || exit 1
elif [ "$cmd" = 'r' ]; then
	revert_relay || exit 1
fi
