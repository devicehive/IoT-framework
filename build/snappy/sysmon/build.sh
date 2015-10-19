#!/bin/bash

DIR="$(dirname "$(readlink -f "$0")")"

go get github.com/shirou/gopsutil/cpu
go get github.com/godbus/dbus
GOOS=linux go build -o $DIR/bin/x86_64/sysmon $DIR/../../../examples/cpu-stats/cpu-stats.go
GOOS=linux GOARCH=arm GOARM=7 go build -o $DIR/bin/armhf/sysmon $DIR/../../../examples/cpu-stats/cpu-stats.go
