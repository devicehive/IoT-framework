# DeviceHive Cloud Gateway D-Bus Daemon

## Decription

`devicehive-cloud` provides D-Bus interface to access DeviceHive cloud server.
It can also be served as a reference implementation of general purpose cloud
conntectivity service (ex: PubNub). It starts as a daemon loading cloud
configuration from `.yml` file. While running it maintains cloud connectivity
and responds to D-Bus API calls from client applications, as well as notifies
applications of incoming messages or status changes.


## Installation
### Ubuntu Snappy Core
Ubuntu Snappy Core runs Raspberry Pi 2, Beagle Bone Black, and a veriety of ARM and
x86 devices, so it can be a good choice for hassle free deployment for a framework.

Steps:
* Install Ubuntu Snappy Core on your favorite system: https://developer.ubuntu.com/en/snappy/start/
* Once Snappy is up and running on your device, install DeviceHive IoT Toolkit:
* Download IoT Toolkit snap from: https://github.com/devicehive/IoT-framework/releases/download/1.0.0-RC1/devicehive-iot-toolkit_1.0.0_multi.snap
* Copy it on your Snappy Device:
```
scp *.snap ubuntu@snappy-host:~
```
* Install snaps using the following command:
```
sudo snappy install devicehive-iot-toolkit_1.0.0_multi.snap  --allow-unauthenticated
```
If any issues occur during snaps install you can check syslog for details:
```
sudo tail -n 100 /var/log/syslog
```

## Configuration

If you are running `devicehive-cloud` as a part of Snappy Framework you can run
`sudo snappy config devicehive-cloud config.yml`, or if you are running it on other
system as a standalone executable, a configuration file can be supplied
with `--conf` command line argument.

See [this](./config.yml) file as an example.


## D-Bus configuration for Ubuntu
In some cases to run `devicehive-cloud` additional system configuration
changes should be made. Need to provide appropriate D-Bus security file
`/etc/dbus-1/system.d/com.devicehive.cloud.conf`.

See [this](./com.devicehive.cloud.conf) file as an example.


## API Reference
TBD


## Building and running it yourself
### How to make a binary?

```
go get github.com/devicehive/IoT-framework/devicehive-cloud
```

### How to run?

```
$GOPATH/bin/devicehive-cloud --conf config.yml
```

### How to make debian package?

```
make debian
```

### How to make snappy package?

```
make snappy
```
