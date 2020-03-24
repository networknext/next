FROM ubuntu:18.04

RUN apt-get -y update && \
    apt install -y wget unzip build-essential make g++ libsparsehash-dev

# libsodium
RUN wget https://github.com/jedisct1/libsodium/releases/download/1.0.17/libsodium-1.0.17.tar.gz && \
    tar xf libsodium-*.tar.gz && \
    cd libsodium-* && \
    ./configure && \
    make && \
    make install && \
    cd .. && \
    ldconfig && \
    rm -rf libsodium-*

# premake5
RUN wget https://github.com/premake/premake-core/releases/download/v5.0.0-alpha14/premake-5.0.0-alpha14-src.zip && \
    unzip premake-*.zip && \
    cd premake-* && \
    cd build/gmake.unix && \
    make && \
    mv ../../bin/release/premake5 /usr/local/bin && \
    cd ../../../ && \
    rm -rf premake-*

COPY . /src/relay

RUN sh /src/relay/build.sh /src/relay

ENTRYPOINT [ "/src/relay/bin/relay" ]
