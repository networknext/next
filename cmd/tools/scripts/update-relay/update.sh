#!/bin/bash

export USERNAME
export IP_ADDRESS
export PUBLIC_KEY
export PRIVATE_KEY
export SSH_KEY
export BINARY

for arg in "$@"
do
    case $arg in
        -u|--user)
        USERNAME="$2"
        shift
        shift
        ;;
        -i|--ipaddress)
        IP_ADDRESS="$2"
        shift
        shift
        ;;
        -p|--publickey)
        PUBLIC_KEY="$2"
        shift
        shift
        ;;
        -s|--privatekey)
        PRIVATE_KEY="$2"
        shift
        shift
        ;;
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

echo Establishing SSH connection to $IP_ADDRESS as $USERNAME
ssh -t -t -i $SSH_KEY $USERNAME@$IP_ADDRESS << EOF
    echo Stopping the relay
    systemctl stop relay

    echo Renaming old relay binary
    cd /app
    mv relay relay-old

    if [ ! -f "/lib/systemd/system/relay.service.backup" ]; then
        echo Creating backup relay.service file relay.service.backup
        cp /lib/systemd/system/relay.service /lib/systemd/system/relay.service.backup
    fi

    echo Closing SSH connection
    exit
EOF

echo Copying relay binary to remote relay
scp -i $SSH_KEY $BINARY $USERNAME@$IP_ADDRESS:/app

echo Pulling down remote relay.service file
scp -i $SSH_KEY $USERNAME@$IP_ADDRESS:/lib/systemd/system/relay.service ~/

echo Adding or replacing env var values
PATCH=""
if grep -q "RELAY_ADDRESS=" ~/relay.service; then
    echo "$(grep "RELAY_ADDRESS=" ~/relay.service | sed -e 's/.*RELAY_ADDRESS=//' -e 's/".*//' | xargs -I RESULT sed -e "s/RESULT/$IP_ADDRESS:40000/" ~/relay.service)" > ~/relay.service
else
    PATCH+="Environment=\"RELAY_ADDRESS=$IP_ADDRESS:40000\"\n"
fi

if grep -q "RELAY_PORT=" ~/relay.service; then
    echo "$(grep "RELAY_PORT=" ~/relay.service | sed -e 's/.*RELAY_PORT=//' -e 's/".*//' | xargs -I RESULT sed -e "s/RESULT/40000/" ~/relay.service)" > ~/relay.service
else
    PATCH+="Environment=\"RELAY_PORT=40000\"\n"
fi

if grep -q "RELAY_BACKEND_HOSTNAME=" ~/relay.service; then
    echo "$(grep "RELAY_BACKEND_HOSTNAME=" ~/relay.service | sed -e 's/.*RELAY_BACKEND_HOSTNAME=//' -e 's/".*//' | xargs -I RESULT sed -e "s/RESULT/relay_backend.dev.spacecats.net:40000/" ~/relay.service)" > ~/relay.service
else
    PATCH+="Environment=\"RELAY_BACKEND_HOSTNAME=35.222.99.199:40000\"\n"
fi

if grep -q "RELAY_PUBLIC_KEY=" ~/relay.service; then
    echo "$(grep "RELAY_PUBLIC_KEY=" ~/relay.service | sed -e 's/.*RELAY_PUBLIC_KEY=//' -e 's/".*//' | xargs -I RESULT sed -e "s:RESULT:$PUBLIC_KEY:" ~/relay.service)" > ~/relay.service
else
    PATCH+="Environment=\"RELAY_PUBLIC_KEY=$PUBLIC_KEY\"\n"
fi

if grep -q "RELAY_PRIVATE_KEY=" ~/relay.service; then
    echo "$(grep "RELAY_PRIVATE_KEY=" ~/relay.service | sed -e 's/.*RELAY_PRIVATE_KEY=//' -e 's/".*//' | xargs -I RESULT sed -e "s:RESULT:$PRIVATE_KEY:" ~/relay.service)" > ~/relay.service
else
    PATCH+="Environment=\"RELAY_PRIVATE_KEY=$PRIVATE_KEY\"\n"
fi

if grep -q "RELAY_ROUTER_PUBLIC_KEY=" ~/relay.service; then
    echo "$(grep "RELAY_ROUTER_PUBLIC_KEY=" ~/relay.service | sed -e 's/.*RELAY_ROUTER_PUBLIC_KEY=//' -e 's/".*//' | xargs -I RESULT sed -e "s:RESULT:SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=:" ~/relay.service)" > ~/relay.service
else
    PATCH+="Environment=\"RELAY_ROUTER_PUBLIC_KEY=SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=\""
fi

# Get the line number that the patch should be inserted
LINE_NUMS=($(grep --line-number '^\s*$' ~/relay.service | sed 's/:.*//'))

if [ -n "$PATCH" ]; then
    sed -i ${LINE_NUMS[1]}"i\\"$PATCH ~/relay.service
fi

echo Copying relay.service file back to remote relay
scp -i $SSH_KEY ~/relay.service $USERNAME@$IP_ADDRESS:/lib/systemd/system/

echo Deleting local relay.service copy
rm ~/relay.service

echo Reestablishing SSH connection to $IP_ADDRESS as $USERNAME
ssh -t -t -i $SSH_KEY $USERNAME@$IP_ADDRESS << EOF
    echo Reloading the daemon
    systemctl daemon-reload

    echo Starting the relay
    systemctl start relay

    echo Closing SSH connection
    exit
EOF
