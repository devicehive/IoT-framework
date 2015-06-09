#!/bin/bash

DIR="$(dirname "$(readlink -f "$0")")"

go get github.com/montanaflynn/stats
go get github.com/godbus/dbus
GOOS=linux go build -o $DIR/bin/x86_64/devicehive-iot-demo $DIR/../../../examples/iot-demo.go
GOOS=linux GOARCH=arm GOARM=7 go build -o $DIR/bin/armhf/devicehive-iot-demo $DIR/../../../examples/iot-demo.go
