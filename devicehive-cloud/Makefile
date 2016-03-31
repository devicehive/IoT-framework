# devicehive-cloud service

ifeq (${VERSION},)
  VERSION=$(shell git describe --tags)
endif
ifeq (${GITHASH},)
  GITHASH=$(shell git log -1 --format='%H')
endif

build: update
	go build -o devicehive-cloud .

install:
	mkdir -p "${DESTDIR}/usr/bin"
	mkdir -p "${DESTDIR}/etc/dbus-1/system.d"
	cp "devicehive-cloud" "${DESTDIR}/usr/bin/"
	cp "com.devicehive.cloud.conf" "${DESTDIR}/etc/dbus-1/system.d/"
	cp "config.yml" "${DESTDIR}/etc/devicehive-cloud.yml"

update:
	go get -v -d -t ./...

clean:
	rm ./devicehive-cloud
#	rm *.deb
#	rm *.snap

DEBIAN_TEMPLATE=debian/template
DEBIAN_PACKAGE_AMD64=devicehive-cloud-${VERSION}_amd64
DEBIAN_PACKAGE_ARM7=devicehive-cloud-${VERSION}_arm
.PHONY: debian debian_amd64 debian_arm7
debian: update debian_amd64 debian_arm7

debian_amd64:
	sudo rm -rf "${DEBIAN_PACKAGE_AMD64}"
	cp -rv "${DEBIAN_TEMPLATE}" "${DEBIAN_PACKAGE_AMD64}"
	GOOS=linux GOARCH=amd64 go build -o "${DEBIAN_PACKAGE_AMD64}/usr/bin/devicehive-cloud"
	mkdir -p "${DEBIAN_PACKAGE_AMD64}/etc/dbus-1/system.d"
	cp "com.devicehive.cloud.conf" "${DEBIAN_PACKAGE_AMD64}/etc/dbus-1/system.d/"
	cp "config.yml" "${DEBIAN_PACKAGE_AMD64}/etc/devicehive-cloud.yml"
	ARCHITECTURE=amd64 VERSION=${VERSION} GITHASH=${GITHASH} envsubst \
		< "${DEBIAN_TEMPLATE}/DEBIAN/control" \
		> "${DEBIAN_PACKAGE_AMD64}/DEBIAN/control"
	sudo chown -R root:root "${DEBIAN_PACKAGE_AMD64}"
	dpkg-deb --build "${DEBIAN_PACKAGE_AMD64}"
	sudo rm -rf "${DEBIAN_PACKAGE_AMD64}"

debian_arm7:
	sudo rm -rf "${DEBIAN_PACKAGE_ARM7}"
	cp -rv "${DEBIAN_TEMPLATE}" "${DEBIAN_PACKAGE_ARM7}"
	GOOS=linux GOARCH=arm GOARM=7 go build -o "${DEBIAN_PACKAGE_ARM7}/usr/bin/devicehive-cloud"
	mkdir -p "${DEBIAN_PACKAGE_ARM7}/etc/dbus-1/system.d"
	cp "com.devicehive.cloud.conf" "${DEBIAN_PACKAGE_ARM7}/etc/dbus-1/system.d/"
	cp "config.yml" "${DEBIAN_PACKAGE_ARM7}/etc/devicehive-cloud.yml"
	ARCHITECTURE=amd64 VERSION=${VERSION} GITHASH=${GITHASH} envsubst \
		< "${DEBIAN_TEMPLATE}/DEBIAN/control" \
		> "${DEBIAN_PACKAGE_ARM7}/DEBIAN/control"
	sudo chown -R root:root "${DEBIAN_PACKAGE_ARM7}"
	dpkg-deb --build "${DEBIAN_PACKAGE_ARM7}"
	sudo rm -rf "${DEBIAN_PACKAGE_ARM7}"


snappy:
	@echo "Snappy not implemented yet"


docker:
	@echo "Docker not implemented yet"

version:
	@echo "Version: ${VERSION}"
	@echo "GitHash: ${GITHASH}"

.PHONY: build install clean snappy docker version
