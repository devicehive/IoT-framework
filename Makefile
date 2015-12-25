all: package

# prepare all daemons
prepare:
	make -C devicehive-cloud prepare
	make -C devicehive-gpio prepare
	make -C devicehive-ble prepare

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

package: snappy_package debian_package archlinux_package
snappy_package:
	make -C build/snappy package

debian_package:
	make -C build/debian package
#	mv -v build/debian/*.deb ./

archlinux_package:
	make -C build/arch package
#	mv -v build/arch/*.pkg.tar.xz ./

clean:
	make -C build/snappy clean
	make -C build/debian clean
	make -C build/arch clean

.PHONY: prepare build install package clean snappy_package debian_package archlinux_package
