# How to build docker images

All `devicehive-cloud` and `devicehive-ble` containers don't have D-Bus installed.
Instead D-Bus daemon from `devicehive-dbus` container is used.

To make all services up use the following command:

```{.sh}
$ cd IoT-framework/docker
$ docker-compose up
```

## Manual run

First of all we need to run D-Bus service (see below how to build images):

```{.sh}
$ docker run -d --name=dbus devicehive/devicehive-dbus:v2
```

Then it's possible to run cloud service:

```{.sh}
$ docker run -d --volumes-from dbus --name=cloud devicehive/devicehive-cloud:v2
```

It's it. Now you can send commands via Admin page.

At the same time it's possible to connect to `dbus` container to monitor D-Bus messages:

```{.sh}
$ docker exec -it dbus /bin/sh
# dbus-monitor --system
```



## Build devicehive-dbus
```{.sh}
$ cd IoT-framework/docker
$ docker build -t devicehive/devicehive-dbus:v2 dbus
```

## Build devicehive-cloud
Please put valid data to `IoT-framework/docker/cloud/config.yml` before building image.

```{.sh}
$ cd IoT-framework/docker
$ docker build -t devicehive/devicehive-cloud:v2 cloud
```

## Build devicehive-ble
```{.sh}
$ cd IoT-framework/docker
$ docker build -t devicehive/devicehive-ble:v2 ble
```
