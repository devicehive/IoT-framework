#!/bin/bash

GOOS=linux go build -o bin/x86_64/sysmon ../../../examples/cpu-stats.go
GOOS=linux GOARCH=arm GOARM=7 go build -o bin/armhf/sysmon ../../../examples/cpu-stats.go
