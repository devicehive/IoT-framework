#!/bin/bash

devicehive-ble/build.sh
devicehive-cloud/build.sh

mkdir /distr
rm /distr/* -rf
cd /distr

snappy build ../devicehive-ble
snappy build ../devicehive-cloud

