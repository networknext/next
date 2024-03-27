
KERNEL = $(shell uname -r)

obj-m += relay_module.o

module: relay_module.c
	@echo building relay module
	sudo rmmod relay_module.ko; echo
	make -C /lib/modules/$(KERNEL)/build M=$(PWD) modules
	sudo modprobe chacha20
	sudo modprobe poly1305
	sudo insmod relay_module.ko; echo

.PHONY: unload
unload:
	