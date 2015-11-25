#!/bin/bash

platform=${PLATFORM:-`uname -i`}   # platform: x86_64|armhf
variant=${VARIANT:-"release"}      # variant: debug|release
echo "platform=$platform variant=$variant"

DIR="$(dirname "$(readlink -f "$0")")"
AJ_PATH=$DIR/../../../devicehive-alljoyn
AJ_CPP=$AJ_PATH/alljoyn/core/alljoyn/build/linux/$platform/$variant/dist/cpp

# Standard+Thin Core
echo "Building thin core ..."
pushd $AJ_PATH/alljoyn
./build-thin-linux-core || exit 1
popd

# Thin Services
echo "Building thin services ..."
pushd $AJ_PATH/alljoyn
./build-thin-linux-services || exit 1
popd

# Go service
echo "Building Go service ..."
pushd $AJ_PATH
case $platform in
    x86_64)
        GOOS=linux go build -o $DIR/bin/$platform/devicehive-alljoyn || exit 1
        ;;
    arm|armhf|RPi|RPi2)
        GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc CXX=arm-linux-gnueabihf-g++ go build -o $DIR/bin/$platform/devicehive-alljoyn || exit 1
        ;;
    *)
        echo "Unsupported platform"
        exit 1
        ;;
esac
popd

echo "Copying alljoyn daemon ..."
mkdir -p $DIR/bin/$platform/
mkdir -p $DIR/lib/$platform/
cp -fv  $AJ_CPP/bin/alljoyn-daemon $DIR/bin/$platform/ || exit 1
cp -fv  $AJ_CPP/lib/liballjoyn.so $DIR/lib/$platform/ || exit 1
cp -fv  $AJ_PATH/alljoyn/core/ajtcl/libajtcl.so $DIR/lib/$platform/ || exit 1

exit 0 # OK

