#!/bin/bash

GOOS=linux go build -o bin/x86_64/devicehive-iot-demo ../../../examples/iot-demo.go
GOOS=linux GOARCH=arm GOARM=7 go build -o bin/armhf/devicehive-iot-demo ../../../examples/iot-demo.go
