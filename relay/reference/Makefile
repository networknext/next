
.PHONY: all
all:	debug

.PHONY: clean
clean:
	rm -f relay

build: *.h *.cpp
	g++ -g *.cpp -lsodium -lcurl -o relay

build-debug: *.h *.cpp
	g++ -DRELAY_DEBUG=1 -g *.cpp -lsodium -lcurl -o relay

.PHONY: debug
debug: build-debug
	RELAY_NAME=debug RELAY_PUBLIC_ADDRESS=192.168.1.40:40000 RELAY_PUBLIC_KEY=XBIAwHE0CppjvPMmtX3TT3QnvmhMbU0KxDpsaFBo+nE= RELAY_PRIVATE_KEY=qlUNbQoc3amdkfNrLflO0Z9/2XJai6zV880AS9cInDc= RELAY_PING_KEY=ANXbw47AaWuu7sidkuw0Cq5cIXU4e8xoqJbSsFC+MT0= RELAY_BACKEND_PUBLIC_KEY=osFv1SPtMkhezNPuLbNbjp/F8ks5I1Y1QVqD0yLd+0o= ./relay
