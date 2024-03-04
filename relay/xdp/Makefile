
#INTERFACE=enp4s0
INTERFACE=lo

KERNEL=6.6.10-76060610-generic

.PHONY: all
all:	local

.PHONY: clean
clean:
	rm -f relay
	rm -f relay_xdp.o

.PHONY: load
load:
	sudo ip link set dev $(INTERFACE) xdp obj relay_xdp.o sec relay_xdp

.PHONY: unload

unload:
	sudo ip link set dev $(INTERFACE) xdp off

.PHONY: local
local: relay_xdp.o relay.c relay_shared.h module relay_xdp.o
	gcc -O2 -g relay.c relay_platform.c relay_base64.c relay_ping_history.c relay_manager.c relay_main.c relay_ping.c relay_bpf.c relay_config.c -o relay -lxdp -lsodium -lcurl /usr/src/linux-headers-$(KERNEL)/tools/bpf/resolve_btfids/libbpf/libbpf.a -lz -lelf
	sudo rm -f /sys/fs/bpf/config_map
	sudo rm -f /sys/fs/bpf/session_map
	sudo rm -f /sys/fs/bpf/whitelist_map
	sudo rm -f /sys/fs/bpf/stats_map
	sudo RELAY_NAME=local.9 RELAY_PUBLIC_ADDRESS=127.0.0.1:2009 RELAY_PUBLIC_KEY=9597P1ZapnmR5X9sTeOLRIE6ZCqGfOEiyJVq2Rb+bV0= RELAY_PRIVATE_KEY=ykNSEqmbzjyz678XfDUnnItB63S1FyBQ7CafO7W1Fgo= RELAY_BACKEND_PUBLIC_KEY=iY1zTFmQASm6ynSSQ1yKihuCrFSqmetrjxGx9Y1xYiA= RELAY_BACKEND_URL=http://127.0.0.1:30000 ./relay

.PHONY: debug
debug: relay_xdp.o relay.c relay_shared.h module
	gcc -DRELAY_DEBUG=1 -DCOMPILE_WITH_BPF=1 -O2 -g relay.c -o relay relay_platform.c relay_base64.c relay_bpf.c relay_config.c relay_debug.c -lxdp -lsodium -lcurl /usr/src/linux-headers-$(KERNEL)/tools/bpf/resolve_btfids/libbpf/libbpf.a -lz -lelf
	sudo rm -f /sys/fs/bpf/stats_map
	sudo rm -f /sys/fs/bpf/state_map
	sudo rm -f /sys/fs/bpf/config_map
	sudo rm -f /sys/fs/bpf/session_map
	sudo rm -f /sys/fs/bpf/whitelist_map
	sudo RELAY_NAME=debug RELAY_PUBLIC_ADDRESS=192.168.1.40:40000 RELAY_PUBLIC_KEY=XBIAwHE0CppjvPMmtX3TT3QnvmhMbU0KxDpsaFBo+nE= RELAY_PRIVATE_KEY=qlUNbQoc3amdkfNrLflO0Z9/2XJai6zV880AS9cInDc= RELAY_PING_KEY=ANXbw47AaWuu7sidkuw0Cq5cIXU4e8xoqJbSsFC+MT0= RELAY_BACKEND_PUBLIC_KEY=osFv1SPtMkhezNPuLbNbjp/F8ks5I1Y1QVqD0yLd+0o= ./relay
	sudo ip link set dev $(INTERFACE) xdp off

relay_xdp.o: relay_xdp.c relay_shared.h
	clang -DRELAY_DEBUG=1 -O2 -g -Ilibbpf/src -target bpf -c relay_xdp.c -o relay_xdp.o

obj-m += relay_module.o

.PHONY: module
module:
	@echo building relay module
	sudo rmmod relay_module.ko; make -C /lib/modules/$(shell uname -r)/build M=$(PWD) modules
	sudo modprobe chacha20
	sudo modprobe poly1305
	sudo insmod relay_module.ko
