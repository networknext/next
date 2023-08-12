#!/bin/sh
sudo journalctl --vacuum-size 10M
sudo NEEDRESTART_SUSPEND=1 apt autoremove -y
sudo NEEDRESTART_SUSPEND=1 apt update -y
sudo NEEDRESTART_SUSPEND=1 apt upgrade -y
sudo NEEDRESTART_SUSPEND=1 apt dist-upgrade -y
sudo NEEDRESTART_SUSPEND=1 apt install libcurl3-gnutls build-essential unattended-upgrades -y
sudo NEEDRESTART_SUSPEND=1 apt autoremove -y
wget https://download.libsodium.org/libsodium/releases/libsodium-1.0.18.tar.gz
tar -zxf libsodium-1.0.18.tar.gz
cd libsodium-1.0.18
./configure
make -j
sudo make install
sudo ldconfig
sudo mkdir /app
