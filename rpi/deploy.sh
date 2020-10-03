#!/bin/bash

# compile
# copy binaries to targets
# generate config files
# generate files
# start tracker
# add files
# start seeders

# OR

# create a folder for each seeder which contains everything(ezl+config, ezs+config)
# copy this folder to the target
# generate the files there? no, because then the seeders will have different files
# then generate them in one place and copy(maybe using ez!) them where needed and if needed

set -e

[[ $# -lt 1 ]] && (echo "missing command argument"; exit 1)

cfg=$(cat config.json)
homePath=$(echo $cfg | jq -r ".homePath")
trackerAddr=$(echo $cfg | jq -r ".trackerAddr")
seeders=$(echo $cfg | jq -r ".seeders")
seedersLength=$(echo $seeders | jq "length")

build() {
    ezDataFmt='{"trackerAddr":"%s"}'
    ezlDataFmt='{"trackerAddr":"%s","seederAddr":"%s","dbPath":"./db"}'
    ezsDataFmt='{"dbPath":"./db"}'

    for i in $(seq 0 $(($seedersLength-1)))
    do
        seeder=$(echo $seeders | jq -r ".[$i]")
        addr=$(echo $seeder | jq -r ".addr")
        isTracker=$(echo $seeder | jq -r ".isTracker")
        isClient=$(echo $seeder | jq -r ".isClient")

        seederDir="seeder_"$addr
        rm -rf $seederDir
        mkdir -p $seederDir

        printf $ezDataFmt $trackerAddr | tpl -t templates/ez.json.tpl > $seederDir/ez.json
        printf $ezlDataFmt $trackerAddr $addr | tpl -t templates/ezl.json.tpl > $seederDir/ezl.json
        printf $ezsDataFmt | tpl -t templates/ezs.json.tpl > $seederDir/ezs.json
    done
}

clean() {
    for i in $(seq 0 $(($seedersLength-1)))
    do
        addr=$(echo $seeders | jq -r ".[$i].addr")
        seederDir="seeder_"$addr
        rm -rf $seederDir
    done
}

case $1 in
    "build")
        build
        ;;
    "clean")
        clean
        ;;
    *)
        echo "unknown command '$1'"
        exit 1
        ;;
esac
