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

# cleanup
apk del --purge go git
rm -rf /var/cache/apk/*
# rm -rf $GOPATH
