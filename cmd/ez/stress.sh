#!/bin/bash

[[ $# -lt 1 ]] && echo "pass number of iterations" && exit 1

for i in $(seq 1 $1)
do
    ./perf.sh run $2
    sleep 5
done
