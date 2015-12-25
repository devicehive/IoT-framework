# select platform
ifdef PLATFORM
  platform=${PLATFORM}
else
  platform=$(shell uname -m)
endif

all: daemon

# build daemon
daemon: daemon_${platform}

daemon_amd64: daemon_x86_64
daemon_x86_64:
	GOOS=linux go build

daemon_arm:
	GOOS=linux GOARCH=arm GOARM=7 go build

# install required Go packages (use godep to update the list)
prepare:
	go get -u github.com/godbus/dbus
	go get -u gopkg.in/fsnotify.v1
	go get -u gopkg.in/yaml.v2

# install daemon to ${DESTDIR}
install: #daemon
	mkdir -p "${DESTDIR}/usr/bin"
	mkdir -p "${DESTDIR}/etc/dbus-1/system.d"
	mkdir -p "${DESTDIR}/usr/share/devicehive-gpio"
	cp devicehive-gpio "${DESTDIR}/usr/bin/"
	cp com.devicehive.gpio.conf "${DESTDIR}/etc/dbus-1/system.d/"
	cp -r profiles "${DESTDIR}/usr/share/devicehive-gpio/"

clean:
	rm -rf devicehive-gpio

.PHONY: prepare daemon install clean
