#!/bin/bash

clean() {
    rm -f cpu.pprof mem.pprof f*B
}

case $1 in
    "clean")
        clean
    ;;
    "run")
        clean
        go build -ldflags "-s -w"
        /usr/bin/time -f "%e,%M" ./ez get -q $2
    ;;
    *)
    ;;
esac
