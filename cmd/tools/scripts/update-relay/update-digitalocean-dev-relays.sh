#!/bin/bash

export SSH_KEY
export BINARY

ips=(
    165.227.10.28
    178.128.227.107
    157.245.214.34
)

public_keys=(
    fYzOGePVAWQrshxFr9sjqzwQwjm927PV6yHIMDU0BUY=
    TrpF9GhyF4N5slGgGLQrSfHaLRGVHPAcOE5BPN3/twU=
    UJlcIgTf1gbWH/YoLebAS6Wig0QKuoVALcug+Xomxio=
)

private_keys=(
    zBlPIdek+TtUOeow93s8V1EG4Nli75sjVgopEWplcgQ=
    QU8lN+/aPu6uTr0xJJCm7j9EX2/vaqM3Z4buBrtebQM=
    hJzdXtvkV7nt8UV9QSCPQPkF5ck09g7pmDzE1FmyfeQ=
)

for arg in "$@"
do
    case $arg in
        -f|--ssh-key)
        SSH_KEY="$2"
        shift
        shift
        ;;
        -b|--binary)
        BINARY="$2"
        shift
        shift
        ;;
    esac
done

if [ -z $SSH_KEY ]; then
    echo "No SSH key provided - use -f or --ssh_key to provide a file path to the SSH key"
    exit
fi

if [ -z $BINARY ]; then
    echo "No relay binary provided - use -b or --binary to provide a file path to the relay binary"
    exit
fi

for i in ${!ips[@]}; do
    echo Updating relay at ip ${ips[$i]}
    ./cmd/tools/scripts/update-relay/update.sh -u root -i ${ips[$i]} -p ${public_keys[$i]} -s ${private_keys[$i]} -f $SSH_KEY -b $BINARY
done
