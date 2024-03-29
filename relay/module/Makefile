
KERNEL = $(shell uname -r)

obj-m += relay_module.o

module: relay_module.c
	@echo building relay module
	make -C /lib/modules/$(KERNEL)/build M=$(shell pwd) modules
	sudo modprobe chacha20
	sudo modprobe poly1305
	sudo insmod relay_module.ko; echo
