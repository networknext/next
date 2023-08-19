#!/bin/sh
if [[ -f /etc/setup_relay_completed ]]; then exit 0; fi
echo sshd: ALL > hosts.deny
echo sshd: $VPN_ADDRESS > hosts.allow
sudo mv hosts.deny /etc/hosts.deny
sudo mv hosts.allow /etc/hosts.allow
sudo touch /etc/setup_relay_has_run
sudo journalctl --vacuum-size 10M
sudo NEEDRESTART_SUSPEND=1 apt autoremove -y
sudo NEEDRESTART_SUSPEND=1 apt update -y
sudo NEEDRESTART_SUSPEND=1 apt upgrade -y
sudo NEEDRESTART_SUSPEND=1 apt dist-upgrade -y
sudo NEEDRESTART_SUSPEND=1 apt install libcurl3-gnutls build-essential -y
sudo NEEDRESTART_SUSPEND=1 apt autoremove -y
wget https://download.libsodium.org/libsodium/releases/libsodium-1.0.18.tar.gz
tar -zxf libsodium-1.0.18.tar.gz
cd libsodium-1.0.18
./configure
make -j
sudo make install
sudo ldconfig
sudo touch /etc/setup_relay_completed
