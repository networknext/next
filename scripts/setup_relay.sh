# input specific to this relay (these change with each relay you set up)

export RELAY_NAME=google.dallas.1
export RELAY_ADDRESS=34.174.121.66:40000
export RELAY_PUBLIC_KEY=dBu7R1Jax6IgnxY87Nhnns1EJue5hs887AAldXTASzI=
export RELAY_PRIVATE_KEY=XsRdnwJ0A+S/tLTElbvD3jjJo7Sx9xgPAFGwNTz+fN4=

# inputs specific to the environment (these change infrequently)

export RELAY_VERSION=2.1.1
export RELAY_BACKEND_HOSTNAME=http://34.117.3.168
export RELAY_ROUTER_PUBLIC_KEY=SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=
export VPN_ADDRESS=45.33.53.242
export ENVIRONMENT=dev

# remove any old journalctl files to free up disk space (if necessary)

sudo journalctl --vacuum-size 200M

# clean up old packages from apt-get to free up disk space (if necessary)

sudo apt autoremove -y

# update installed packages

sudo apt update -y
sudo apt upgrade -y
sudo apt dist-upgrade -y
sudo apt autoremove -y

# install build essentials so we can build libsodium

sudo apt install build-essential -y

# install unattended upgrades so the relay keeps up to date with security fixes

sudo apt install unattended-upgrades -y

# only allow ssh from vpn address

echo sshd: ALL > hosts.deny
echo sshd: $VPN_ADDRESS > hosts.allow
sudo mv hosts.deny /etc/hosts.deny
sudo mv hosts.allow /etc/hosts.allow

# make the relay prompt cool

sudo echo "export PS1=\"\[\033[36m\]$RELAY_NAME [$ENVIRONMENT] \[\033[00m\]\w # \"" >> ~/.bashrc
sudo echo "source ~/.bashrc" >> ~/.profile.sh

# build and install libsodium optimized for this relay

wget https://download.libsodium.org/libsodium/releases/libsodium-1.0.18.tar.gz
tar -zxf libsodium-1.0.18.tar.gz
cd libsodium-1.0.18
./configure
make -j
sudo make install
ldconfig
cd ~

# download the relay binary and rename it to 'relay'

wget https://storage.googleapis.com/relay_artifacts/relay-$RELAY_VERSION
sudo mv relay-$RELAY_VERSION relay
sudo chmod +x relay

# setup the relay environment file

sudo cat > relay.env <<- EOM
RELAY_BACKEND_HOSTNAME=$RELAY_BACKEND_HOSTNAME
RELAY_PUBLIC_KEY=$RELAY_PUBLIC_KEY
RELAY_PRIVATE_KEY=$RELAY_PRIVATE_KEY
RELAY_ROUTER_PUBLIC_KEY=$RELAY_ROUTER_PUBLIC_KEY
RELAY_ADDRESS=$RELAY_ADDRESS
EOM

# setup the relay service file

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

# install the relay service, then start it and watch the logs

sudo systemctl enable /app/relay.service
sudo systemctl start relay
sudo journalctl -fu relay -n 100
