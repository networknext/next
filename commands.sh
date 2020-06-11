#!/bin/sh
PWD=`pwd`
echo "pwd: $PWD" 
# echo "cmd line arg: $1"
if [ "$1" = 1 ] 
then 
	redis-cli flushall && make BACKEND_LOG_LEVEL=info dev-relay-backend
elif [ "$1" = 2 ]
then
	make dev-multi-relays
elif [ "$1" = 3 ]
then
	make BACKEND_LOG_LEVEL=info dev-server-backend
elif [ "$1" = 4 ]
then
	make dev-server
elif [ "$1" = 5 ]
then
	make dev-multi-clients
elif [ "$1" = 6 ]
then
	make JWT_AUDIENCE="oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n" dev-portal
else 
	echo "no command executed"
fi
