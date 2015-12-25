# select platform
ifdef PLATFORM
  platform=${PLATFORM}
else
  platform=$(shell uname -m)
endif


all: daemon

daemon: daemon_${platform}

daemon_x86_64:
	GOOS=linux go build

daemon_arm:
	GOOS=linux GOARCH=arm GOARM=7 go build

clean:
	rm -rf devicehive-cloud

install: daemon
	mkdir -p ${DESTDIR}/usr/bin
	mkdir -p ${DESTDIR}/etc/dbus-1/system.d/
	cp devicehive-cloud ${DESTDIR}/usr/bin/
	cp exampleconf.yml ${DESTDIR}/etc/devicehive-cloud.conf
	cp com.devicehive.cloud.conf ${DESTDIR}/etc/dbus-1/system.d/

.PHONY: clean
