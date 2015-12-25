all: snappy_package debian_package archlinux_package

# build all daemons
build:
	make -C devicehive-cloud daemon
	make -C devicehive-gpio daemon
	make -C devicehive-ble daemon

# install all daemon to $DESTDIR
install:
	make -C devicehive-cloud install
	make -C devicehive-gpio install
	make -C devicehive-ble install

snappy_package:
	make -C build/snappy package

debian_package:
	make -C build/debian package

archlinux_package:
	make -C build/arch package
#	mv -v build/arch/*.pkg.tar.xz ./

clean:
	make -C build/snappy clean
	make -C build/debian clean
	make -C build/arch clean

.PHONY: clean snappy_package debian_package archlinux_package
