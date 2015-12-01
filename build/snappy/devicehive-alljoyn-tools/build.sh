#!/bin/bash

platform=${PLATFORM:-`uname -i`}   # platform: x86_64|armhf
variant=${VARIANT:-"release"}      # variant: debug|release
echo "platform=$platform variant=$variant"

DIR="$(dirname "$(readlink -f "$0")")"
OUT_DIR=$DIR/../devicehive-alljoyn
AJ_PATH=$DIR/../../../devicehive-alljoyn
AJ_CPP=$AJ_PATH/alljoyn/core/alljoyn/build/linux/$platform/$variant/dist/cpp

mkdir -p $OUT_DIR/bin/$platform/
mkdir -p $OUT_DIR/lib/$platform/

# ensure libssl.a & libcap.so are available on ARM
case $platform in
    arm|armhf|RPi|RPi2|BBB)
        [ -d $AJ_PATH/alljoyn/core/alljoyn/build_core/conf/linux/armhf ] || patch -d $AJ_PATH/alljoyn/core/alljoyn -p1 -i $DIR/alljoyn-armhf.patch
        [ -e /usr/arm-linux-gnueabihf/lib/libcap.so ] || sudo ln -s $DIR/armhf/libcap.so /usr/arm-linux-gnueabihf/lib/libcap.so
        [ -e /usr/arm-linux-gnueabihf/lib/libssl.a ] || sudo ln -s $DIR/armhf/libssl.a /usr/arm-linux-gnueabihf/lib/libssl.a
        # TODO: remove these symbolic links at the end
        ;;
esac

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
        GOOS=linux go build -o $OUT_DIR/bin/$platform/devicehive-alljoyn || exit 1
        ;;
    arm|armhf|RPi|RPi2)
        GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc CXX=arm-linux-gnueabihf-g++ go build -o $OUT_DIR/bin/$platform/devicehive-alljoyn || exit 1
        ;;
    *)
        echo "Unsupported platform"
        exit 1
        ;;
esac
popd

echo "Copying alljoyn daemon ..."
cp -fv  $AJ_CPP/bin/alljoyn-daemon $OUT_DIR/bin/$platform/ || exit 1
cp -fv  $AJ_CPP/lib/liballjoyn.so $OUT_DIR/lib/$platform/ || exit 1
cp -fv  $AJ_PATH/alljoyn/core/ajtcl/libajtcl.so $OUT_DIR/lib/$platform/ || exit 1

exit 0 # OK

