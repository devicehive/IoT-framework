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
	go get -u github.com/gorilla/websocket
	go get -u gopkg.in/yaml.v2
	go get -u github.com/devicehive/devicehive-go/devicehive

# install daemon to ${DESTDIR}
install: #daemon
	mkdir -p "${DESTDIR}/usr/bin"
	mkdir -p "${DESTDIR}/etc/dbus-1/system.d"
	cp devicehive-cloud "${DESTDIR}/usr/bin/"
	cp exampleconf.yml "${DESTDIR}/etc/devicehive-cloud.conf"
	cp com.devicehive.cloud.conf "${DESTDIR}/etc/dbus-1/system.d/"

clean:
	rm -rf devicehive-cloud

.PHONY: prepare daemon install clean
