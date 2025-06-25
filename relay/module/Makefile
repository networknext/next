
KERNEL = $(shell uname -r)

obj-m += relay_module.o

module: relay_module.c
	@echo building relay module
	rm -f *.o
	rm -f *.ko
	make -C /lib/modules/$(KERNEL)/build M=$(shell pwd) modules
	sudo modprobe chacha20
	sudo modprobe poly1305
	sudo rmmod -f relay_module
	sudo insmod relay_module.ko; echo
