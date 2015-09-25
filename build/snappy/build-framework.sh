#!/bin/bash

echo "Building framework..."
devicehive-framework/build.sh

echo "snappy Building package (devicehive-framework) ..."
snappy build devicehive-framework
