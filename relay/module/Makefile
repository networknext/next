
KERNEL = $(shell uname -r)

obj-m += relay_module.o

module: relay_module.c
	@echo building relay module
	sudo cp /sys/kernel/btf/vmlinux /usr/lib/modules/`uname -r`/build/
	rm -f *.o
	rm -f *.ko
	make -C /lib/modules/$(KERNEL)/build M=$(shell pwd) modules
	sudo modprobe chacha20
	sudo modprobe poly1305
	sudo rmmod -f relay_module >/dev/null 2>&1; echo
	sudo insmod relay_module.ko >/dev/null 2>&1; echo
