#!/bin/bash

DIR="$(dirname "$(readlink -f "$0")")"

go get github.com/godbus/dbus

GOOS=linux go build -o $DIR/bin/x86_64/heart-pulse-demo $DIR/../../../examples/heart-pulse-demo.go
GOOS=linux GOARCH=arm GOARM=7 go build -o $DIR/bin/armhf/heart-pulse-demo $DIR/../../../examples/heart-pulse-demo.go
