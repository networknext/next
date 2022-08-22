
mkdir dist

parallel ::: \
'cd ./dist && g++ -I../sdk4/include -shared -o sdk4.so ../sdk4/source/*.cpp -lsodium -lcurl -lpthread -lm -framework CoreFoundation -framework SystemConfiguration' \
'cd ./dist && g++ -I../sdk5/include -shared -o sdk5.so ../sdk5/source/*.cpp -lsodium -lcurl -lpthread -lm -framework CoreFoundation -framework SystemConfiguration' \
'cd ./dist && g++ -o reference_relay ../reference/relay/*.cpp -lsodium -lcurl -lpthread -lm -framework CoreFoundation -framework SystemConfiguration' \
'go build -o ./dist/func_tests_backend ./cmd/func_tests_backend/*.go' \
'go build -o ./dist/magic_backend ./cmd/magic_backend/*.go' \
'go build -o ./dist/magic_frontend ./cmd/magic_frontend/*.go' \
'go build -o ./dist/relay_gateway ./cmd/relay_gateway/*.go' \
'go build -o ./dist/relay_backend ./cmd/relay_backend/*.go' \
'go build -o ./dist/relay_frontend ./cmd/relay_frontend/*.go' \
