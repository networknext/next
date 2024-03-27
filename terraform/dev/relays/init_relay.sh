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
sudo NEEDRESTART_SUSPEND=1 apt install libcurl3-gnutls-dev build-essential vim wget libsodium-dev flex bison -y
sudo NEEDRESTART_SUSPEND=1 apt autoremove -y

# setup for xdp/bpf
sudo NEEDRESTART_SUSPEND=1 apt install clang linux-headers-generic linux-headers-`uname -r` unzip libc6-dev-i386 gcc-12 dwarves libelf-dev pkg-config m4 libpcap-dev net-tools -y
sudo cp /sys/kernel/btf/vmlinux /usr/lib/modules/`uname -r`/build/
wget https://github.com/xdp-project/xdp-tools/releases/download/v1.4.2/xdp-tools-1.4.2.tar.gz
tar -zxf xdp-tools-1.4.2.tar.gz
cd xdp-tools-1.4.2
./configure
make -j && sudo make install
cd lib/libbpf/src
make -j && sudo make install
sudo ldconfig
cd /

# install relay module
sudo mkdir /relay_module
cd /relay_module
sudo wget https://storage.googleapis.com/xdp_network_next_relay_artifacts/relay_module.tar.gz
sudo tar -zxf relay_module.tar.gz
sudo make
sudo mkdir /lib/modules/`uname -r`/kernel/net/relay_module
sudo mv relay_module.ko /lib/modules/`uname -r`/kernel/net/relay_module
echo "chacha20" > modules.txt
echo "poly1305" >> modules.txt
echo "relay_module" >> modules.txt
sudo mv modules.txt /etc/modules
sudo depmod

sudo touch /etc/init_relay_completed
