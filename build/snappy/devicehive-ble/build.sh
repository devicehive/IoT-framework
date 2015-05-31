#!/bin/bash

DIR="$(dirname "$(readlink -f "$0")")"

go get github.com/devicehive/gatt
go get github.com/godbus/dbus
GOOS=linux go build -o $DIR/bin/x86_64/devicehive-ble $DIR/../../../devicehive-ble/devicehive-ble.go
GOOS=linux GOARCH=arm GOARM=7 go build -o $DIR/bin/armhf/devicehive-ble $DIR/../../../devicehive-ble/devicehive-ble.go
