#!/bin/bash


DIR="$(dirname "$(readlink -f "$0")")"

go get github.com/godbus/dbus
go get github.com/gorilla/websocket
go get github.com/mibori/gopencils
GOOS=linux go build -o $DIR/bin/x86_64/devicehive-cloud $DIR/../../../devicehive-cloud/devicehive-cloud.go $DIR/../../../devicehive-cloud/rest-serve.go $DIR/../../../devicehive-cloud/ws-serve.go
GOOS=linux GOARCH=arm GOARM=7 go build -o $DIR/bin/armhf/devicehive-cloud $DIR/../../../devicehive-cloud/devicehive-cloud.go $DIR/../../../devicehive-cloud/rest-serve.go $DIR/../../../devicehive-cloud/ws-serve.go
