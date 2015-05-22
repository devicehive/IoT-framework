#!/bin/bash

GOOS=linux go build -o bin/x86_64/devicehive-cloud ../../../devicehive-cloud/devicehive-cloud.go ../../../devicehive-cloud/rest-serve.go ../../../devicehive-cloud/ws-serve.go
GOOS=linux GOARCH=arm GOARM=7 go build -o bin/armhf/devicehive-cloud ../../../devicehive-cloud/devicehive-cloud.go ../../../devicehive-cloud/rest-serve.go ../../../devicehive-cloud/ws-serve.go
