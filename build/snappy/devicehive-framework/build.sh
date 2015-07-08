#!/bin/bash

DIR="$(dirname "$(readlink -f "$0")")"

# Common
go get github.com/godbus/dbus

# Bluetooth LE
go get github.com/devicehive/gatt
GOOS=linux go build -o $DIR/bin/x86_64/devicehive-ble $DIR/../../../devicehive-ble/devicehive-ble.go
GOOS=linux GOARCH=arm GOARM=7 go build -o $DIR/bin/armhf/devicehive-ble $DIR/../../../devicehive-ble/devicehive-ble.go

# Cloud
go get gopkg.in/yaml.v2
go get github.com/gorilla/websocket
go get github.com/mibori/gopencils
GOOS=linux go build -o $DIR/bin/x86_64/devicehive-cloud $DIR/../../../devicehive-cloud/devicehive-cloud.go $DIR/../../../devicehive-cloud/rest-serve.go $DIR/../../../devicehive-cloud/ws-serve.go
GOOS=linux GOARCH=arm GOARM=7 go build -o $DIR/bin/armhf/devicehive-cloud $DIR/../../../devicehive-cloud/devicehive-cloud.go $DIR/../../../devicehive-cloud/rest-serve.go $DIR/../../../devicehive-cloud/ws-serve.go

# Enocean
cp -f $DIR/../../../devicehive-enocean/enocean-daemon $DIR/bin/devicehive-enocean
PYTHONUSERBASE=$DIR pip3 install --user enocean

# GPIO
cp -f  $DIR/../../../devicehive-gpio/gpio-daemon $DIR/bin/devicehive-gpio
cp -fr $DIR/../../../devicehive-gpio/profiles $DIR/bin/


