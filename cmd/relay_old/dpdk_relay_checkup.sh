#!/bin/bash

#get dpdk
echo "Getting DPDK and dependencies!!!"
sudo apt-get install -y nano git sysstat wget make g++ dh-autoreconf pkg-config cmake iptables unzip software-properties-common libcurl4-openssl-dev pkg-config build-essential autoconf automake uuid-dev chrony pciutils libnuma-dev linux-headers-generic linux-headers-`uname -r` python openssh-server && \
wget https://fast.dpdk.org/rel/dpdk-18.02.tar.xz && \
tar xvf dpdk-18.02.tar.xz && \
cd dpdk-18.02
export RTE_SDK="$(pwd)"
export RTE_TARGET=x86_64-native-linuxapp-gcc

#setup dpdk, install drivers
echo "Setup DPDK, install drivers!!!"
cd $RTE_SDK && \
make install T=$RTE_TARGET && \
cd $RTE_TARGET && \
sudo modprobe uio && \
sudo insmod kmod/igb_uio.ko

#bind current intel capable nics
echo "Binding current DPDK capable NICs!!!"
cd $RTE_SDK && \
./usertools/dpdk-devbind.py --status | grep ixgbe > /tmp/foo;
while read line ; do
    PCI="$(echo $line | cut -c-12)"
    echo "Binding $PCI to igb_uio for DPDK!"
    sudo ./usertools/dpdk-devbind.py --bind=igb_uio $PCI
done < /tmp/foo

#setup hugetables before every app, teardown after each app is done
echo "Setup hugetables!!!"
sudo sh -c "echo 2048 >  /sys/devices/system/node/node0/hugepages/hugepages-2048kB/nr_hugepages" && \
sudo mkdir -p /mnt/huge && \
sudo mount -t hugetlbfs nodev /mnt/huge

#run hello world app
echo "Running HelloWorld APP!!!"
cd $RTE_SDK/examples/helloworld && \
make && \
sudo build/helloworld -l 0-7 -m 2
sudo umount /mnt/huge

#setup hugetables before every app, teardown after each app is done
echo "Setup hugetables!!!"
sudo sh -c "echo 2048 >  /sys/devices/system/node/node0/hugepages/hugepages-2048kB/nr_hugepages" && \
sudo mkdir -p /mnt/huge && \
sudo mount -t hugetlbfs nodev /mnt/huge

#run fwding app
echo "Running Basic Fwd APP!!!"
cd $RTE_SDK/$RTE_TARGET/build/app/test-pmd && \
echo "($PWD)"
sudo ./testpmd -c f -n 4 -- -i
sudo umount /mnt/huge
