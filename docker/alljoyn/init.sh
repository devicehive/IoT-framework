#!/bin/sh -e

export GOPATH=/go

BRANCH=master
REPO_BASE=https://github.com/devicehive
AJ_TAG=v14.12
AJ_REPO_BASE=https://git.allseenalliance.org/gerrit
AJ_VARIANT=release
AJ_CPU=`uname -m`
case "$AJ_CPU" in
	armv7l)
		export CROSS_COMPILE=
		AJ_CPU=arm
		;;
esac

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
	--alljoyn=*)
		AJ_TAG="${1#*=}"
		shift
		;;
	--alljoyn)
		AJ_TAG="$2"
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
apk add --update go git scons g++ libcap-dev openssl-dev libstdc++ libgcc

# cloning repositories
mkdir -p $GOPATH/src/github.com/devicehive
cd $GOPATH/src/github.com/devicehive
# git clone -b $BRANCH --single-branch $REPO_BASE/devicehive-go
git clone -b $BRANCH --single-branch $REPO_BASE/IoT-framework

# get alljoyn
BASEDIR=$GOPATH/src/github.com/devicehive/IoT-framework/devicehive-alljoyn
cd $BASEDIR
git clone $AJ_REPO_BASE/core/alljoyn.git alljoyn/core/alljoyn
git clone $AJ_REPO_BASE/core/ajtcl.git alljoyn/core/ajtcl
git clone $AJ_REPO_BASE/services/base_tcl.git alljoyn/services/base_tcl

# build core/alljoyn
cd $BASEDIR/alljoyn/core/alljoyn
git checkout -b $AJ_TAG $AJ_TAG # tag
patch -p1 -i /tmp/alljoyn_no_empty_cross.patch # CROSS_COMPILE required for arm
patch -p1 -i /tmp/alljoyn_swap.patch # fix swapXX functions
echo "building core/alljoyn..."
scons BINDINGS=cpp WS=off BT=off ICE=off VARIANT=$AJ_VARIANT CPU=$AJ_CPU >/tmp/alljoyn.log 2>&1

# build core/ajtcl
cd $BASEDIR/alljoyn/core/ajtcl
git checkout -b $AJ_TAG $AJ_TAG # tag
patch -p1 -i /tmp/ajtcl_va_copy.patch # fix missing __va_copy macro
echo "building core/ajtcl..."
scons BINDINGS=cpp WS=off BT=off ICE=off VARIANT=$AJ_VARIANT CPU=$AJ_CPU >/tmp/ajtcl.log 2>&1

# build services/base_tcl
cd $BASEDIR/alljoyn/services/base_tcl
git checkout -b $AJ_TAG $AJ_TAG # tag
#echo "building services/base_tcl..."
#scons BINDINGS=cpp WS=off BT=off ICE=off VARIANT=$AJ_VARIANT CPU=$AJ_CPU >/tmp/base_tcl.log 2>&1

# update sources and build
cd $BASEDIR
# `go get -u` doesn't work with branches - it tryes to update from master!
go get -d -t -v ./...
go build -o /usr/bin/devicehive-alljoyn

# copy configuration file for D-Bus, it will be processed by /start.sh script
mkdir -p /usr/share/dbus-1/system.d && \
	cp com.devicehive.alljoyn.conf /usr/share/dbus-1/system.d/

# copy alljoyn daemon and libraries
AJ_DIST=$BASEDIR/alljoyn/core/alljoyn/build/linux/$AJ_CPU/$AJ_VARIANT/dist
cp $AJ_DIST/cpp/bin/alljoyn-daemon /usr/bin/
cp $AJ_DIST/cpp/lib/liballjoyn.so /usr/lib/
cp $BASEDIR/alljoyn/core/ajtcl/libajtcl.so /usr/lib/

# cleanup
apk del --purge go git scons g++ libcap-dev openssl-dev
rm -rf /var/cache/apk/*
rm -rf $GOPATH
