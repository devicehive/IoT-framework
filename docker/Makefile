# make PREFIX=devicehive VERSION=2.0 ARCH=armhf

# prefix
ifeq (${PREFIX},)
  prefix=
else
  prefix=${PREFIX}/
endif

# version
ifeq (${VERSION},)
  version=2.0
else
  version=${VERSION}
endif

# architecture
ifeq (${ARCH},)
  # auto-detection...
  arch=$(shell uname -m)
  ifeq (${arch},$(filter ${arch},x86_64 amd64))
    # default one, no prefix
    dash_arch=
    dot_arch=
  else ifeq (${arch},$(filter ${arch},armv7l))
    # armhf prefix
    dash_arch=-armhf
    dot_arch=.armhf
  endif
else
  arch=${ARCH}
  dash_arch=-${arch}
  dot_arch=.${arch}
endif

all: build

# build
build_dbus:
	docker build -t ${prefix}devicehive-dbus:${version}${dash_arch} -f dbus/Dockerfile${dot_arch} dbus
build_cloud:
	docker build -t ${prefix}devicehive-cloud:${version}${dash_arch} -f cloud/Dockerfile${dot_arch} cloud
build_ble:
	docker build -t ${prefix}devicehive-ble:${version}${dash_arch} -f ble/Dockerfile${dot_arch} ble
build_examples:
	docker build -t ${prefix}devicehive-examples:${version}${dash_arch} -f examples/Dockerfile${dot_arch} examples
build: build_dbus build_cloud build_ble build_examples
.PHONY: build build_dbus build_cloud build_ble build_examples

# push
push_dbus: build_dbus
	docker push ${prefix}devicehive-dbus:${version}${dash_arch}
push_cloud: build_cloud
	docker push ${prefix}devicehive-cloud:${version}${dash_arch}
push_ble: build_ble
	docker push ${prefix}devicehive-ble:${version}${dash_arch}
push_examples: build_examples
	docker push ${prefix}devicehive-examples:${version}${dash_arch}
push: push_dbus push_cloud push_ble push_examples
.PHONY: push push_dbus push_cloud push_ble push_examples

up:
	docker-compose -f docker-compose${dot_arch}.yml up
.PHONY: up

# WARNING: remove *ALL* containers/images
clean_docker: clean_docker_containers clean_docker_images
clean_docker_containers:
	docker rm -f $(shell docker ps -a -q)
clean_docker_images:
	docker rmi -f $(shell docker images -q)
.PHONY: clean_docker clean_docker_containers clean_docker_images
