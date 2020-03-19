#!/bin/bash

# give some names to the IPs

SSH_RELAY_LINODE_FREMONT='root@45.79.228.25'
SSH_RELAY_LINODE_ATLANTA='root@172.105.158.138'
SSH_RELAY_DIGITALOCEAN_SANFRANCISCO='root@165.227.10.28'
SSH_RELAY_DIGITALOCEAN_TORONTO='root@178.128.227.107'
SSH_RELAY_DIGITALOCEAN_NEWYORK='root@157.245.214.34'
SSH_RELAY_VULTR_ATLANTA='root@144.202.31.156'
SSH_RELAY_VULTR_DALLAS='root@45.76.58.249'
SSH_RELAY_VULTR_LOSANGELES='root@45.63.51.158'
SSH_RELAY_VULTR_NEWYORK='root@149.28.53.151'
SSH_RELAY_VULTR_SANJOSE='root@149.28.213.194'
SSH_RELAY_VULTR_SEATTLE='root@149.248.32.18'
SSH_RELAY_GOOGLE_LOSANGELES_1='root@34.94.28.121'
SSH_RELAY_GOOGLE_LOSANGELES_2='root@35.236.75.236'
SSH_RELAY_GOOGLE_SOUTHCAROLINA='root@34.73.237.48'
SSH_RELAY_MAXIHOST_CHICAGO='root@186.233.186.30'
SSH_RELAY_AMAZON_OHIO='ubuntu@3.134.174.102'
SSH_RELAY_AMAZON_VIRGINIA='ubuntu@34.199.100.144'

# enable the desired relays here

ssh_targets=(
"$SSH_RELAY_LINODE_FREMONT"
"$SSH_RELAY_LINODE_ATLANTA"
"$SSH_RELAY_DIGITALOCEAN_SANFRANCISCO"
"$SSH_RELAY_DIGITALOCEAN_TORONTO"
"$SSH_RELAY_DIGITALOCEAN_NEWYORK"
"$SSH_RELAY_VULTR_ATLANTA"
"$SSH_RELAY_VULTR_DALLAS"
"$SSH_RELAY_VULTR_LOSANGELES"
"$SSH_RELAY_VULTR_NEWYORK"
"$SSH_RELAY_VULTR_SANJOSE"
"$SSH_RELAY_VULTR_SEATTLE"
"$SSH_RELAY_GOOGLE_LOSANGELES_1"
"$SSH_RELAY_GOOGLE_LOSANGELES_2"
"$SSH_RELAY_GOOGLE_SOUTHCAROLINA"
"$SSH_RELAY_MAXIHOST_CHICAGO"
)

ssh_targets=(
"$SSH_RELAY_LINODE_FREMONT"
)

troublesome_amazon_targets=(
"$SSH_RELAY_AMAZON_OHIO"
"$SSH_RELAY_AMAZON_VIRGINIA"
)

troublesome_amazon_targets=()

print_usage() {
	printf "i [identiy file]\tSame as ssh's -i flag\n\n"
	printf "b [binary]\t\tThe binary file you want to send to the relays\n\n"
	printf "y [yes]\t\tDon't ask to proceed, you know what you're doing\n\n"
}

ssh_key="$HOME/.ssh/relay.dev.key"
binary="./cmd/relay_old/bin/relay"
proceed=''
while getopts 'i:b:yh' flag; do
	case "${flag}" in
		i) 
			ssh_key="${OPTARG}" 
			;;
		b)
			binary="${OPTARG}"
			;;
		y)
			proceed='y'
			;;
		h) 
			print_usage
			exit 1
			;;
		*)
			print_usage
			exit 1
			;;
	esac
done

echo "Using relay key located at $ssh_key"
echo "Sending the file $binary"

# in case the wrong binary was specified 
# or this was executed by accident 
# or you're just not feeling like it

if [ -z "$proceed" ]; then
	printf "Proceed? (y/n): "

	if read line; then
		proceed="$line"
	fi

	if [ ! "$proceed" == 'y' ]; then
		echo "Exiting"
		exit 1
	fi
fi

# most of the updates occure in this loop

for target in "${ssh_targets[@]}"; do
	echo "Stopping the relay service on $target"
	ssh -i "$ssh_key" "$target" '$(which bash)' << EOF
	systemctl stop relay
EOF

echo "Sending the binary"
scp -i "$ssh_key" "$binary" "$target:/app/relay"

echo "Starting the relay service"
ssh -i "$ssh_key" "$target" '$(which bash)' << EOF
	systemctl start relay
EOF

done

# but amazon has a slightly different routine because you're not the root user

for target in "${troublesome_amazon_targets[@]}"; do
	echo "Stopping the relay service on $target"
	ssh -i "$ssh_key" "$target" '$(which bash)' << EOF
	sudo systemctl stop relay
EOF

echo "Sending the binary"
scp -i "$ssh_key" "$binary" "$target:~/relay"

echo "Starting the relay service"
ssh -i "$ssh_key" "$target" '$(which bash)' << EOF
	cd /app
	sudo mv ~/relay ./
	sudo systemctl start relay
EOF

done

