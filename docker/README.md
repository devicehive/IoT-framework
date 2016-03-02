# How to build docker images

All `devicehive-cloud` and `devicehive-ble` containers don't have D-Bus installed.
Instead D-Bus daemon from `devicehive-dbus` container is used.

To make all services up use the following command:

```{.sh}
cd IoT-framework/docker
docker-compose up
```

## Manual run

First we need to run D-Bus service:

```{.sh}
(host)$ docker run -it --name dbus --entrypoint /bin/sh devicehive/devicehive-dbus:v2
(dbus)# /start.sh
(dbus)# dbus-monitor --system # to catch what is going on D-Bus
```

Then it's possible to run cloud service:

```{.sh}
(host)$ docker run -it --volumes-from dbus --name cloud --entrypoint /bin/sh devicehive/devicehive-dbus:v2
(cloud)# /start.sh
```

It's it. Now you can send commands via Admin page.


## Build devicehive-cloud

Please put valid data to `IoT-framework/docker/cloud/config.yml` before building image.

```{.sh}
cd IoT-framework/docker
docker build -t devicehive/devicehive-cloud:v2 cloud
```

## Build devicehive-ble
```{.sh}
cd IoT-framework/docker
docker build -t devicehive/devicehive-ble:v2 ble
```

## Build devicehive-dbus
```{.sh}
cd IoT-framework/docker
docker build -t devicehive/devicehive-dbus:v2 dbus
```
