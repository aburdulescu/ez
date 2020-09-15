#!/bin/bash

set -e

if [[ $# -lt 1 ]]
then
    echo "error: need file name"
    exit 1
fi

if [[ $# -lt 2 ]]
then
    echo "error: need file size(MB)"
    exit 1
fi

dd if=/dev/urandom of=$1 bs=1M count=$2 status=progress
