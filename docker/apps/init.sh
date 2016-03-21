#!/bin/sh -e

export GOPATH=/go

REPO_BASE=https://github.com/devicehive
BRANCH=master

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
apk add --update go git python py-dbus py-gobject

# cloning repositories
mkdir -p $GOPATH/src/github.com/devicehive
cd $GOPATH/src/github.com/devicehive
git clone -b $BRANCH --single-branch $REPO_BASE/devicehive-go
git clone -b $BRANCH --single-branch $REPO_BASE/IoT-framework

# cleanup
# apk del --purge go git
rm -rf /var/cache/apk/*
# rm -rf $GOPATH
