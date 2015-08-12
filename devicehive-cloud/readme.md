#DeviceHive Cloud Gateway D-Bus Daemon 

##Decription
Devicehive-cloud provides d-bus interface to access DeviceHive cloud server. It can also serves as a refrence implementation of general purpose cloud conntectivity service for (ex: PubNub). As a binary it starts as a daemon loading cloud configuration from .yml file. While running it maintains cloud connectivity and responds to d-bus API calls from client applications, as well as notifies applications of incoming messages or status changes. 

## Installation
### Ubuntu Snappy Core
Ubuntu Snappy Core runs Raspberry Pi 2, Beagle Bone Black, and a veriety of ARM and x86 devices, so it can be a good choice for hassle free deployment for a frameowrk.

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
If you are running devicehive-cloud as a part of Snappy Framework you can run ```sudo snappy config devicehive-cloud config.yml```, or if you are running it on other system as a standalone binary, a configuration file can be supplied in ```--conf``` parameter. 

Sample config:
```
URL: http://52.1.250.210:8080/dh/rest
AccessKey: 1jwKgLYi/CdfBTI9KByfYxwyQ6HUIEfnGSgakdpFjgk=

DeviceID: 0B24431A-EC99-4887-8B4F-38C3CEAF1D03
DeviceName: snappy-go-gateway

SendNotificatonQueueCapacity: 2047	# Optional: default value is 2048
LoggingLevel: verbose               # Optional: can be 'info', 'verbose', 'debug'
                                    # Default: 'info'
```

## API Reference
TBD

## Building and running it yourself
###How to make a binary?
```
go get github.com/devicehive/IoT-framework/tree/master/devicehive-cloud
cd $GOPATH/src/github.com/devicehive/IoT-framework/tree/master/devicehive-cloud
go install
```
###How to run?
```
$GOPATH/bin/devicehive-cloud -conf=deviceconf.yml
```
