#!/bin/bash

set -e
set -x

# you also need these packages installed:
# sudo apt install build-essentials gdb libsparsehash-dev

if [ "$(which premake5)" == "" ]; then
    cd ~
    rm -Rf ~/premake-* || true
    rm -f premake5 || true
    curl -sSL -o premake-linux.tar.gz https://github.com/premake/premake-core/releases/download/v5.0.0-alpha14/premake-5.0.0-alpha14-linux.tar.gz
    tar -xvf premake-linux.tar.gz
    rm premake-linux.tar.gz
    chmod a+x premake5
    sudo mv premake5 /usr/local/bin/premake5
fi

if [ ! -f /usr/local/include/sodium.h ]; then
    cd ~
    rm -Rf ~/libsodium-* || true
    wget https://github.com/jedisct1/libsodium/releases/download/1.0.17/libsodium-1.0.17.tar.gz
    tar xf libsodium-*.tar.gz
    rm libsodium-*.tar.gz
    cd libsodium-*
    ./configure
    make
    sudo make install
    cd ..
    sudo ldconfig
    rm -rf libsodium-*
fi

cd "${0%/*}"
cd ..
premake5 gmake && make -j32