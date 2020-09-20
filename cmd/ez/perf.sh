#!/bin/bash

clean() {
    rm -f perf.cpu f1GB
}

case $1 in
    "clean")
        clean
    ;;
    "run")
        clean
        go build -ldflags "-s -w"
        /usr/bin/time -f "%E %MKB" ./ez get $2
    ;;
    *)
    ;;
esac
