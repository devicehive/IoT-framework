#!/bin/sh -e

export GOPATH=/go

BRANCH=master
REPO_BASE=https://github.com/devicehive

# parse command line
while [ $# -gt 0 ]; do
	case "$1" in
	--branch=*)
		BRANCH="${1#*=}"
		shift
		;;
	--branch)
		BRANCH="$2"
		shift 2
		;;
	--extra-repo=*)
		echo "${1#*=}" >> /etc/apk/repositories
		shift
		;;
	--extra-repo)
		echo "$2" >> /etc/apk/repositories
		shift 2
		;;
	*)
		echo "'$1' is unknown option, ignored"
		shift
	esac
done

# install dependencies
apk add --update go git bluez

# cloning repositories
mkdir -p $GOPATH/src/github.com/devicehive
cd $GOPATH/src/github.com/devicehive
#git clone -b $BRANCH --single-branch $REPO_BASE/devicehive-go
git clone -b $BRANCH --single-branch $REPO_BASE/IoT-framework

# update sources and build
cd $GOPATH/src/github.com/devicehive/IoT-framework/devicehive-ble
# `go get -u` doesn't work with branches - it tryes to update from master!
go get -d -t -v ./...
go build -o /usr/bin/devicehive-ble

# copy configuration file for D-Bus, it will be processed by start.sh script
mkdir -p /usr/share/dbus-1/system.d && \
	cp com.devicehive.bluetooth.conf /usr/share/dbus-1/system.d/

# cleanup
apk del --purge go git
rm -rf /var/cache/apk/*
rm -rf $GOPATH
