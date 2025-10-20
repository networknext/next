#!/bin/bash

# if we have not finished initializing the relay, wait for it!

while [[ ! -f /etc/init_relay_completed ]]
do
  echo "waiting for relay init to finish..."
  sleep 10
done

# apt update and upgrade is sometimes necessary

sudo apt update -y && sudo apt upgrade -y

# IMPORTANT: if we are not running a 6.5 kernel, upgrade the kernel. we need ubuntu 22.04 LTS with linux kernel 6.5 for xdp relay to work

major=$(uname -r | awk -F '.' '{print $1}')
minor=$(uname -r | awk -F '.' '{print $2}')

echo linux kernel version is $major.$minor

if [[ $major -lt 6 ]]; then
  echo "upgrading linux kernel to 6.5... please run setup again on this relay after it reboots"
  sudo DEBIAN_FRONTEND=noninteractive NEEDRESTART_SUSPEND=1 apt install linux-generic-hwe-22.04 -y
  sudo reboot
fi

# make the relay prompt cool

echo making the relay prompt cool

sudo echo "export PS1=\"\[\033[36m\]$RELAY_NAME [$ENVIRONMENT] \[\033[00m\]\w # \"" >> ~/.bashrc
sudo echo "source ~/.bashrc" >> ~/.profile.sh

# download the relay binary and rename it to 'relay'

echo downloading relay binary

rm -f $RELAY_VERSION

wget https://storage.googleapis.com/$RELAY_ARTIFACTS_BUCKET_NAME/$RELAY_VERSION --no-cache

if [ ! $? -eq 0 ]; then
    echo "download relay binary failed"
    exit 1
fi

sudo mv $RELAY_VERSION relay

sudo chmod +x relay

# setup the relay environment file

echo setting up relay environment

sudo cat > relay.env <<- EOM
RELAY_NAME=$RELAY_NAME
RELAY_PUBLIC_ADDRESS=$RELAY_PUBLIC_ADDRESS
RELAY_INTERNAL_ADDRESS=$RELAY_INTERNAL_ADDRESS
RELAY_PUBLIC_KEY=$RELAY_PUBLIC_KEY
RELAY_PRIVATE_KEY=$RELAY_PRIVATE_KEY
RELAY_BACKEND_URL=$RELAY_BACKEND_URL
RELAY_BACKEND_PUBLIC_KEY=$RELAY_BACKEND_PUBLIC_KEY
EOM

# if we need to reboot, it's best to do it now before we try to install linux headers because the kernel version may change

if [ -f /var/run/reboot-required ]; then
    echo "rebooting. please run setup again on this relay"
    sudo reboot
fi

# setup linux tools, headers and vmlinux BTF file needed for bpf. this requires 6.5+ linux kernel to work

sudo NEEDRESTART_SUSPEND=1 apt install dwarves linux-headers-`uname -r` linux-tools-`uname -r` -y

sudo cp /sys/kernel/btf/vmlinux /usr/lib/modules/`uname -r`/build/

# install relay module

sudo rm -rf ~/relay_module
mkdir -p ~/relay_module
cd ~/relay_module
wget https://storage.googleapis.com/next_network_next_relay_artifacts/relay_module.tar.gz
tar -zxf relay_module.tar.gz
make
sudo mkdir -p /lib/modules/`uname -r`/kernel/net/relay_module
sudo mv relay_module.ko /lib/modules/`uname -r`/kernel/net/relay_module

# setup relay module to load on reboot

cd ~
echo "chacha20" > modules.txt
echo "poly1305" >> modules.txt
echo "relay_module" >> modules.txt
sudo mv modules.txt /etc/modules
sudo depmod

# setup the relay service file

echo setting up relay service file

sudo cat > relay.service <<- EOM
[Unit]
Description=Network Next Relay
ConditionPathExists=/app/relay
After=network.target

[Service]
Type=simple
LimitNOFILE=1024
WorkingDirectory=/app
ExecStart=/app/relay
EnvironmentFile=/app/relay.env
Restart=always
RestartSec=10
TimeoutStopSec=90s
[Install]
WantedBy=multi-user.target
EOM

# move everything into the /app dir

echo moving everything into /app

sudo rm -rf /app
sudo mkdir /app
sudo mv relay /app/relay
sudo mv relay.env /app/relay.env
sudo mv relay.service /app/relay.service

# limit maximum journalctl logs to 200MB so we don't run out of disk space

echo limiting max journalctl logs to 200MB

sudo sed -i "s/\(.*SystemMaxUse= *\).*/\SystemMaxUse=200M/" /etc/systemd/journald.conf
sudo systemctl restart systemd-journald

# install the relay service, then start it and watch the logs

echo installing relay service

sudo systemctl enable /app/relay.service

echo starting relay service

sudo systemctl start relay

sudo touch /etc/setup_relay_completed

echo setup completed

sudo reboot
