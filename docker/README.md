# How to build docker images

All `devicehive-cloud` and `devicehive-ble` containers don't have D-Bus installed.
Instead D-Bus daemon from `devicehive-dbus` container is used.

## Build devicehive-cloud
```{.sh}
cd IoT-framework/docker
docker build -t devicehive/devicehive-cloud:v2 -f Dockerfile.cloud .
```

## Build devicehive-ble
```{.sh}
cd IoT-framework/docker
docker build -t devicehive/devicehive-ble:v2 -f Dockerfile.ble .
```

## Build devicehive-dbus
```{.sh}
cd IoT-framework/docker
docker build -t devicehive/devicehive-dbus:v2 -f Dockerfile.dbus .
```
