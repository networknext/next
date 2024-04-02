#!/bin/sh

if test -f /etc/init_relay_completed; then
  echo "init relay completed"
  exit 0
fi

sudo touch /etc/init_relay_has_run

# only allow ssh from the vpn address

echo sshd: ALL > hosts.deny
echo sshd: $VPN_ADDRESS > hosts.allow
sudo mv hosts.deny /etc/hosts.deny
sudo mv hosts.allow /etc/hosts.allow

# clean out any old journalctl logs so we have space to do stuff

sudo journalctl --vacuum-size 10M

# install necessary packages

sudo NEEDRESTART_SUSPEND=1 apt autoremove -y
sudo NEEDRESTART_SUSPEND=1 apt update -y
sudo NEEDRESTART_SUSPEND=1 apt upgrade -y
sudo NEEDRESTART_SUSPEND=1 apt dist-upgrade -y
sudo NEEDRESTART_SUSPEND=1 apt full-upgrade -y
sudo NEEDRESTART_SUSPEND=1 apt install libcurl3-gnutls-dev build-essential vim wget libsodium-dev flex bison clang unzip libc6-dev-i386 gcc-12 dwarves libelf-dev pkg-config m4 libpcap-dev net-tools -y
sudo NEEDRESTART_SUSPEND=1 apt autoremove -y

# install libxdp and libbpf from source. this is neccessary for the xdp relay to work

cd ~
wget https://github.com/xdp-project/xdp-tools/releases/download/v1.4.2/xdp-tools-1.4.2.tar.gz
tar -zxf xdp-tools-1.4.2.tar.gz
cd xdp-tools-1.4.2
./configure
make -j && sudo make install

cd lib/libbpf/src
make -j && sudo make install
sudo ldconfig
cd /

sudo touch /etc/init_relay_completed

# IMPORTANT: if we are not running a 6.5 kernel, upgrade the kernel. we need ubuntu 22.04 LTS with linux kernel 6.5 for xdp relay to work

if [[ ! `uname -r` == "6.5."* ]]; then
  echo "upgrading linux kernel to 6.5... please run setup again on this relay after it reboots"
  sudo NEEDRESTART_SUSPEND=1 apt install linux-generic-hwe-22.04 -y
  sudo reboot  
else
  sudo NEEDRESTART_SUSPEND=1 apt install linux-headers-`uname -r` linux-tools-`uname -r` -y # saves time in setup_relay.sh
fi
