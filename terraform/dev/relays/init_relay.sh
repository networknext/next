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

# setup relay with basic software needed
sudo journalctl --vacuum-size 10M
sudo NEEDRESTART_SUSPEND=1 apt autoremove -y
sudo NEEDRESTART_SUSPEND=1 apt update -y
sudo NEEDRESTART_SUSPEND=1 apt upgrade -y
sudo NEEDRESTART_SUSPEND=1 apt dist-upgrade -y
sudo NEEDRESTART_SUSPEND=1 apt full-upgrade -y
sudo NEEDRESTART_SUSPEND=1 apt install libcurl3-gnutls-dev build-essential vim wget libsodium-dev flex bison clang unzip libc6-dev-i386 gcc-12 dwarves libelf-dev pkg-config m4 libpcap-dev net-tools -y
sudo NEEDRESTART_SUSPEND=1 apt autoremove -y



sudo touch /etc/init_relay_completed
