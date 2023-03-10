#!/bin/sh

ENVIRONMENT              = $1
RELAY_NAME               = $2
RELAY_PUBLIC_ADDRESS     = $3
RELAY_INTERNAL_ADDRESS   = $4
RELAY_PUBLIC_KEY         = $5
RELAY_PRIVATE_KEY        = $6
RELAY_BACKEND_HOSTNAME   = $7
RELAY_BACKEND_PUBLIC_KEY = $8
RELAY_VERSION            = $9
RELAY_ARTIFACTS_BUCKET   = $10
VPN_ADDRESS              = $11

# remove any old journalctl files to free up disk space (if necessary)

sudo journalctl --vacuum-size 10M

# clean up old packages from apt-get to free up disk space (if necessary)

sudo NEEDRESTART_SUSPEND=1 apt apt autoremove -y

# update installed packages

sudo NEEDRESTART_SUSPEND=1 apt apt update -y
sudo NEEDRESTART_SUSPEND=1 apt apt upgrade -y
sudo NEEDRESTART_SUSPEND=1 apt dist-upgrade -y
sudo NEEDRESTART_SUSPEND=1 apt apt autoremove -y

# we need libcurl

sudo apt install libcurl3-gnutls -y
sudo ldconfig

# install build essentials so we can build libsodium

sudo apt install build-essential -y

# install unattended upgrades so the relay keeps up to date with security fixes

sudo apt install unattended-upgrades -y

# only allow ssh from vpn address

echo sshd: ALL > hosts.deny
echo sshd: $VPN_ADDRESS > hosts.allow
sudo mv hosts.deny /etc/hosts.deny
sudo mv hosts.allow /etc/hosts.allow

# make the relay prompt super cool

sudo echo "export PS1=\"\[\033[36m\]$RELAY_NAME [$ENVIRONMENT] \[\033[00m\]\w # \"" >> ~/.bashrc
sudo echo "source ~/.bashrc" >> ~/.profile.sh

# build and install libsodium optimized for this relay

wget https://download.libsodium.org/libsodium/releases/libsodium-1.0.18.tar.gz
tar -zxf libsodium-1.0.18.tar.gz
cd libsodium-1.0.18
./configure
make -j
sudo make install
sudo ldconfig
cd ~

# download the relay binary and rename it to 'relay'

wget https://storage.googleapis.com/$RELAY_ARTIFACTS_BUCKET/relay-$RELAY_VERSION --no-cache
mv relay-$RELAY_VERSION relay
chmod +x relay

# setup the relay environment file

cat > relay.env <<- EOM
RELAY_NAME=$RELAY_NAME
RELAY_PUBLIC_ADDRESS=$RELAY_PUBLIC_ADDRESS
RELAY_INTERNAL_ADDRESS=$RELAY_INTERNAL_ADDRESS
RELAY_PUBLIC_KEY=$RELAY_PUBLIC_KEY
RELAY_PRIVATE_KEY=$RELAY_PRIVATE_KEY
RELAY_BACKEND_HOSTNAME=$RELAY_BACKEND_HOSTNAME
RELAY_BACKEND_PUBLIC_KEY=$RELAY_BACKEND_PUBLIC_KEY
EOM

# setup the relay service file

cat > relay.service <<- EOM
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
Restart=on-failure
RestartSec=12

[Install]
WantedBy=multi-user.target
EOM

# move everything into the /app dir

sudo rm -rf /app
sudo mkdir /app
sudo mv relay /app/relay
sudo mv relay.env /app/relay.env
sudo mv relay.service /app/relay.service

# limit maximum journalctl logs to 200MB so we don't run out of disk space

sudo sed -i "s/\(.*SystemMaxUse= *\).*/\SystemMaxUse=200M/" /etc/systemd/journald.conf
sudo systemctl restart systemd-journald

# install the relay service so it starts on boot

sudo systemctl enable /app/relay.service

# start the relay service

sudo systemctl start relay
