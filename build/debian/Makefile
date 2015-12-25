# select platform
ifdef PLATFORM
  platform=${PLATFORM}
else
  platform=$(shell uname -m)
endif

# DEBIAN: replace "x86_64" with "amd64"
ifeq "${platform}" "x86_64"
  platform=amd64
endif

template=devicehive-cloud.template
version=0.1
githash=$(shell git rev-parse HEAD)
debian=devicehive-IoT_${version}-${githash}_${platform}
destdir=$(shell pwd)/pkg/${debian}

all: package

# create package
package:
	rm -rf "${destdir}"
	mkdir -p "${destdir}"
	cp -r "${template}/." "${destdir}/"
	make -C ../.. build
	DESTDIR="${destdir}" make -C ../.. install
	ARCHITECTURE=${platform} VERSION=${version} GITHASH=${githash} \
	  envsubst < "${template}/DEBIAN/control" > "${destdir}/DEBIAN/control"
	dpkg-deb --build "${destdir}" "${debian}.deb"
#	rm -rf "${destdir}"


# clean packages
clean:
	rm -rf *.deb

.PHONY: package clean
