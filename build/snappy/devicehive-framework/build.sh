#!/bin/bash

DIR="$(dirname "$(readlink -f "$0")")"

# Common
echo "Fetching github.com/godbus/dbus ..."
go get github.com/godbus/dbus

# Bluetooth LE
echo "Fetching github.com/devicehive/gatt"
go get github.com/devicehive/gatt

echo "Building devicehive-ble (amd64) ..."
GOOS=linux go build -o $DIR/bin/x86_64/devicehive-ble $DIR/../../../devicehive-ble/devicehive-ble.go

echo "Building devicehive-ble (arm) ..."
GOOS=linux GOARCH=arm GOARM=7 go build -o $DIR/bin/armhf/devicehive-ble $DIR/../../../devicehive-ble/devicehive-ble.go

# Cloud
echo "Fetching gopkg.in/yaml.v2 ..."
go get gopkg.in/yaml.v2

echo "Fetching github.com/gorilla/websocket ..."
go get github.com/gorilla/websocket
# go get github.com/mibori/gopencils
# GOOS=linux go build -o $DIR/bin/x86_64/devicehive-cloud $DIR/../../../devicehive-cloud/devicehive-cloud.go $DIR/../../../devicehive-cloud/rest-serve.go $DIR/../../../devicehive-cloud/ws-serve.go
# GOOS=linux GOARCH=arm GOARM=7 go build -o $DIR/bin/armhf/devicehive-cloud $DIR/../../../devicehive-cloud/devicehive-cloud.go $DIR/../../../devicehive-cloud/rest-serve.go $DIR/../../../devicehive-cloud/ws-serve.go

echo "Building devicehive-daemon (amd64) ..."
GOOS=linux go build -o $DIR/bin/x86_64/devicehive-cloud $DIR/../../../devicehive-cloud/*.go
echo "Building devicehive-cloud (arm) ..."
GOOS=linux GOARCH=arm GOARM=7 go build -o $DIR/bin/armhf/devicehive-cloud $DIR/../../../devicehive-cloud/*.go


# Enocean
echo "Copying enocean-daemon ..."
cp -fv $DIR/../../../devicehive-enocean/enocean-daemon $DIR/bin/devicehive-enocean

echo "Pip3 install enocean ..."
PYTHONUSERBASE=$DIR pip3 install --user enocean

# GPIO

echo "Copying gpio-daemon ..."
cp -fv  $DIR/../../../devicehive-gpio/gpio-daemon $DIR/bin/devicehive-gpio

echo "Copying gpio-daemon profiles ..."
cp -frv $DIR/../../../devicehive-gpio/profiles $DIR/bin/


