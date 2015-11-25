#!/bin/bash

echo "building AllJoyn ..."
PLATFORM=x86_64 devicehive-alljoyn/build.sh || exit 1
PLATFORM=armhf devicehive-alljoyn/build.sh || exit 1

echo "building snappy package (devicehive-alljoyn) ..."
snappy build devicehive-alljoyn || exit 1

exit 0 # OK

