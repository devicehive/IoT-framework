# devicehive-cloud service

ifeq (${VERSION},)
  VERSION=$(shell git describe --tags)
endif
ifeq (${GITHASH},)
  GITHASH=$(shell git log -1 --format='%H')
endif

build:
	go build .

intall:
	go install .

update:
	@echo TODO: go get -v -u -d -t ./...

clean:
	rm ./devicehive-cloud

DEBIAN_TEMPLATE=debian/template
DEBIAN_PACKAGE_AMD64=devicehive-cloud-${VERSION}_amd64
DEBIAN_PACKAGE_ARM7=devicehive-cloud-${VERSION}_arm
.PHONY: debian debian_amd64 debian_arm7
debian: update debian_amd64 debian_arm7
# TODO: provide init scripts!

debian_amd64:
	rm -rf "${DEBIAN_PACKAGE_AMD64}"
	cp -rv "${DEBIAN_TEMPLATE}" "${DEBIAN_PACKAGE_AMD64}"
	GOOS=linux GOARCH=amd64 go build -o "${DEBIAN_PACKAGE_AMD64}/usr/bin/devicehive-cloud"
	mkdir -p "${DEBIAN_PACKAGE_AMD64}/etc/dbus-1/system.d"
	cp "com.devicehive.cloud.conf" "${DEBIAN_PACKAGE_AMD64}/etc/dbus-1/system.d/"
	cp "config.yml" "${DEBIAN_PACKAGE_AMD64}/etc/devicehive-cloud.yml"
	ARCHITECTURE=amd64 VERSION=${VERSION} GITHASH=${GITHASH} envsubst \
		< "${DEBIAN_TEMPLATE}/DEBIAN/control" \
		> "${DEBIAN_PACKAGE_AMD64}/DEBIAN/control"
	dpkg-deb --build "${DEBIAN_PACKAGE_AMD64}"
	rm -rf "${DEBIAN_PACKAGE_AMD64}"

debian_arm7:
	rm -rf "${DEBIAN_PACKAGE_ARM7}"
	cp -rv "${DEBIAN_TEMPLATE}" "${DEBIAN_PACKAGE_ARM7}"
	GOOS=linux GOARCH=arm GOARM=7 go build -o "${DEBIAN_PACKAGE_ARM7}/usr/bin/devicehive-cloud"
	mkdir -p "${DEBIAN_PACKAGE_ARM7}/etc/dbus-1/system.d"
	cp "com.devicehive.cloud.conf" "${DEBIAN_PACKAGE_ARM7}/etc/dbus-1/system.d/"
	cp "config.yml" "${DEBIAN_PACKAGE_ARM7}/etc/devicehive-cloud.yml"
	ARCHITECTURE=amd64 VERSION=${VERSION} GITHASH=${GITHASH} envsubst \
		< "${DEBIAN_TEMPLATE}/DEBIAN/control" \
		> "${DEBIAN_PACKAGE_ARM7}/DEBIAN/control"
	dpkg-deb --build "${DEBIAN_PACKAGE_ARM7}"
	rm -rf "${DEBIAN_PACKAGE_ARM7}"


snappy:
	@echo "Snappy not implemented yet"

docker:
	@echo "Docker not implemented yet"

version:
	@echo "Version: ${VERSION}"
	@echo "GitHash: ${GITHASH}"

.PHONY: build install clean snappy docker version
