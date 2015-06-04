#DeviceHive Cloud Gateway D-Bus Daemon 
##How to make a binary?

```
go get github.com/devicehive/IoT-framework/tree/master/devicehive-cloud
cd $GOPATH/src/github.com/devicehive/IoT-framework/tree/master/devicehive-cloud
go install
```


##How to run?
```
$GOPATH/bin/devicehive-cloud -conf=deviceconf.yml
```

##Configuaration file
Example:
```
URL: http://52.1.250.210:8080/dh/rest
AccessKey: 1jwKgLYi/CdfBTI9KByfYxwyQ6HUIEfnGSgakdpFjgk=

DeviceID: 0B24431A-EC99-4887-8B4F-38C3CEAF1D03
DeviceName: snappy-go-gateway

SendNotificatonQueueCapacity: 2047	# Optional: default value is 2048
LoggingLevel: verbose               # Optional: can be 'info', 'verbose', 'debug'
                                    # Default: 'info'
```
