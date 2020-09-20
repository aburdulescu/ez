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
        /usr/bin/time -f "%E %MKB" ./ez get sha1-a4b4b44140b5ca06ee075e21ca002eea287587bb
    ;;
    *)
    ;;
esac
