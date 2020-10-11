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

# Don't use config file for ezs, use flags
# Generate service files for ezt and ezs

set -e

[[ $# -lt 1 ]] && (echo "missing command argument"; exit 1)

cfg=$(cat config.json)
homePath=$(echo $cfg | jq -r ".homePath")
trackerAddr=$(echo $cfg | jq -r ".trackerAddr")
seeders=$(echo $cfg | jq -r ".seeders")
seedersLength=$(echo $seeders | jq "length")

build() {
    serviceDataFmt='{"description":"%s","execStart":"%s"}'

    pushd ../
    make clean
    GOOS=linux GOARCH=arm make
    popd

    for i in $(seq 0 $(($seedersLength-1)))
    do
        seeder=$(echo $seeders | jq -r ".[$i]")
        addr=$(echo $seeder | jq -r ".addr")
        isTracker=$(echo $seeder | jq -r ".isTracker")
        isClient=$(echo $seeder | jq -r ".isClient")

        seederDir="seeder_"$addr
        rm -rf $seederDir
        mkdir -p $seederDir

        cp ../cmd/ezl/ezl $seederDir/ezl

        cp ../cmd/ezs/ezs $seederDir/ezs
        execStart="$homePath/ezs -dbpath="$(dirname $homePath)"/db -seedaddr="$addr" -trackeraddr="$trackerAddr
        printf $serviceDataFmt "ez seeder server" "$execStart" | tpl -t templates/service.tpl > $seederDir/ezs.service

        if [[ $isTracker == "true" ]]
        then
            cp ../cmd/ezt/ezt $seederDir/ezt
            printf $serviceDataFmt "ez tracker server" $homePath/ezt | tpl -t templates/service.tpl > $seederDir/ezt.service
        fi
        if [[ $isClient == "true" ]]
        then
            cp ../cmd/ez/ez $seederDir/ez
        fi
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

deploy() {
    if [[ -z $RPI_PASSWORD ]]
    then
        echo "define the RPI_PASSWORD variable"
        exit 1
    fi
    for i in $(seq 0 $(($seedersLength-1)))
    do
        addr=$(echo $seeders | jq -r ".[$i].addr")
        seederDir="seeder_"$addr
        echo "deploy $seederDir"
        sshpass -p $RPI_PASSWORD ssh pi@$addr rm -rf $homePath
        sshpass -p $RPI_PASSWORD ssh pi@$addr mkdir -p $homePath
        sshpass -p $RPI_PASSWORD scp $seederDir/* pi@$addr:$homePath/
    done
}

case $1 in
    "build")
        build
        ;;
    "clean")
        clean
        ;;
    "deploy")
        deploy
        ;;
    *)
        echo "unknown command '$1'"
        exit 1
        ;;
esac
