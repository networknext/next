FROM ubuntu:bionic AS builder


RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        build-essential \
        libsodium-dev \
        unzip \
        curl \
        ca-certificates

ARG PREMAKE_VERSION=5.0.0-alpha14

RUN cd /tmp && \
    curl -fsSLO https://github.com/premake/premake-core/releases/download/v${PREMAKE_VERSION}/premake-${PREMAKE_VERSION}-src.zip && \
    unzip premake-5.0.0-alpha14-src.zip && \
    cd premake-5.0.0-alpha14 && \
    cd build/gmake.unix && \
    make && \
    mv ../../bin/release/premake5 /usr/local/bin && \
    cd / && \
    rm -rf /tmp/premake-*

WORKDIR /src

ADD . .

ARG CONFIG=release_static64

RUN premake5 gmake && make -j config=${CONFIG}

FROM ubuntu:bionic

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y -q --no-install-recommends libsodium23 && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /src/bin/server /bin/
COPY --from=builder /src/bin/client /bin/


CMD ["/bin/server"]
