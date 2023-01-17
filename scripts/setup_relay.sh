# input specific to this relay (these change with each relay you set up)

export RELAY_NAME=google.saopaulo.1
export RELAY_ADDRESS=34.151.248.241:40000
export RELAY_PUBLIC_KEY=qunlVxGncMg5b650wXgtYBmJAzetry+K9ancBayMWzw=
export RELAY_PRIVATE_KEY=1vpJ9L6jntr+KvqHSkZvgH9EnkVE/stS+60pfAdXEkg=

# inputs specific to the environment (these change infrequently)

export RELAY_VERSION=2.1.0
export RELAY_BACKEND_HOSTNAME=http://34.117.3.168
export RELAY_ROUTER_PUBLIC_KEY=SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=
export VPN_ADDRESS=45.33.53.242
export ENVIRONMENT=dev

# only allow ssh from vpn address

sudo rm -f /etc/hosts.deny
sudo echo sshd: ALL > /etc/hosts.deny

sudo rm -f /etc/hosts.allow
sudo echo sshd: $VPN_ADDRESS > /etc/hosts.allow

# make the relay prompt cool

sudo echo "export PS1=\"\[\033[36m\]$RELAY_NAME [$ENVIRONMENT] \[\033[00m\]\w # \"" >> ~/.bashrc

# build and install libsodium optimized for this relay

wget https://download.libsodium.org/libsodium/releases/libsodium-1.0.18.tar.gz
tar -zxf libsodium-1.0.18.tar.gz
cd libsodium-1.0.18
./configure
make -j
sudo make install
ldconfig
cd ~

# download the relay binary and copy it to the app directory

wget https://storage.googleapis.com/relay_artifacts/relay-$RELAY_VERSION
sudo rm -rf /app
sudo mkdir /app
mv relay-$RELAY_VERSION /app/relay
chmod +x /app/relay

# setup the relay environment file

cat > /app/relay.env <<- EOM
RELAY_BACKEND_HOSTNAME=$RELAY_BACKEND_HOSTNAME
RELAY_PUBLIC_KEY=$RELAY_PUBLIC_KEY
RELAY_PRIVATE_KEY=$RELAY_PRIVATE_KEY
RELAY_ROUTER_PUBLIC_KEY=$RELAY_ROUTER_PUBLIC_KEY
RELAY_ADDRESS=$RELAY_ADDRESS
EOM

# setup the relay service file

cat > /app/relay.service <<- EOM
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

# limit maximum journalctl logs to 200MB so we don't run out of disk space

sudo journalctl --vacuum-size 200M
sudo sed -i "s/\(.*SystemMaxUse= *\).*/\SystemMaxUse=200M/" /etc/systemd/journald.conf
sudo systemctl restart systemd-journald

# install the relay service, then start it and watch the logs

sudo systemctl enable /app/relay.service
sudo systemctl start relay
sudo journalctl -fu relay -n 100
