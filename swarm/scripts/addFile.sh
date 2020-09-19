#!/bin/bash

mkdir -p /go-ez/files
cd /go-ez/files
/go-ez/bin/mkf.sh $1 $2
cd /go-ez/config/
/go-ez/bin/ezl add /go-ez/files/$1
