#!/bin/sh -e

BRANCH=${1:-master}
REPO_BASE=https://github.com/devicehive
export GOPATH=/go

# install dependencies
apk add --update go git

# cloning repositories
mkdir -p $GOPATH/src/github.com/devicehive
cd $GOPATH/src/github.com/devicehive
git clone -b $BRANCH --single-branch $REPO_BASE/devicehive-go
git clone -b $BRANCH --single-branch $REPO_BASE/IoT-framework

# update sources and build
cd $GOPATH/src/github.com/devicehive/IoT-framework/devicehive-cloud
# `go get -u` doesn't work with branches - it tryes to update from master!
go get -d -t -v ./...
go build -o /usr/bin/devicehive-cloud

# copy configuration file for D-Bus, it will be processed by start.sh script
mkdir -p /usr/share/dbus-1/system.d && \
	cp com.devicehive.cloud.conf /usr/share/dbus-1/system.d/

# cleanup
apk del --purge go git
rm -rf /var/cache/apk/*
rm -rf $GOPATH
