#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# The root of the build/dist directory
HARNSGATEWAY_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"

echo "go detail version: $(go version)"

goversion=$(go version |awk -F ' ' '{printf $3}' |sed 's/go//g')

echo "go version: $goversion"

X=$(echo $goversion|awk -F '.' '{printf $1}')
Y=$(echo $goversion|awk -F '.' '{printf $2}')

if [ $X -lt 1 ] ; then
	echo "go major version must >= 1, now is $X"
	exit 1
fi

if [ $Y -lt 20 ] ; then
	echo "go minor version must >= 20, now is $Y"
	exit 1
fi
