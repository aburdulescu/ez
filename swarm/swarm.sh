#!/bin/bash

set -e

cfg=$(cat swarm.json)

subnet_name=$(echo $cfg | jq -r ".subnet.name")
subnet_ip=$(echo $cfg | jq -r ".subnet.ip")
subnet_mask=$(echo $cfg | jq -r ".subnet.mask")
subnet_ip_prefix=$(echo $subnet_ip|cut -d "." -f "1,2,3")"."

num_seeders=$(echo $cfg | jq -r ".numSeeders")

seeder_dbpath=$(echo $cfg | jq -r ".seederDbPath")

files=$(echo $cfg | jq -r ".files")

tracker_ip=$subnet_ip_prefix"254"

files_dir=$(pwd)/files
swarm_dir=$(pwd)


build_subnet() {
    docker network inspect $subnet_name 1>>/dev/null 2>>/dev/null || docker network create --subnet "$subnet_ip$subnet_mask" $subnet_name
}

build_tracker() {
    docker build -t ez_tracker -f dockerfiles/Dockerfile.tracker .
}

build_files() {
    rm -rf $files_dir
    mkdir -p $files_dir
    cd $files_dir
    files_length=$(echo $files | jq "length")
    for i in $(seq 0 $(($files_length-1)))
    do
        name=$(echo $files | jq -r ".[$i].name")
        size=$(echo $files | jq -r ".[$i].size")
        $swarm_dir/scripts/mkf.sh $name $size
    done
    cd -
}

build_seeders() {
    seeder_entrypoint_fmt="{\"files\":"$files",\"seederDbPath\":\"%s\",\"seedAddr\":\"%s\",\"trackerAddr\":\"%s\"}"

    for i in $(seq 1 $num_seeders)
    do
        ip_suffix=$((i+1))

        seedAddr=$subnet_ip_prefix$ip_suffix":22201"

        rm -f seeder-entrypoint.sh

        printf "$seeder_entrypoint_fmt" $seeder_dbpath $seedAddr $tracker_ip | tpl -t templates/seeder-entrypoint.sh.tpl > seeder-entrypoint.sh

        chmod +x seeder-entrypoint.sh

        docker build -t ez_seeder_$i -f dockerfiles/Dockerfile.seeder .
    done

    rm seeder-entrypoint.sh
}

build() {
    build_subnet
    build_tracker
    build_files
    build_seeders
}

run() {
    ip=$1
    img=$2
    echo $ip $img
    docker run \
           --rm \
           -d \
           -v $files_dir:/ez/files \
           --network $subnet_name \
           --ip $ip \
           --name $img \
           $img
}

clean() {
    docker image rm -f ez_tracker
    for i in $(seq 1 $num_seeders)
    do
        docker image rm -f ez_seeder_$i
    done
    rm -rf $files_dir
}

start() {
    run $tracker_ip ez_tracker
    for i in $(seq 1 $num_seeders)
    do
        ip_suffix=$((i+1))
        run $subnet_ip_prefix$ip_suffix ez_seeder_$i
    done
}

stop() {
    docker stop ez_tracker
    for i in $(seq 1 $num_seeders)
    do
        docker stop ez_seeder_$i
    done
}

[[ $# -lt 1 ]] && (echo "missing command argument"; exit 1)

case $1 in
    "build")
        build
        ;;
    "clean")
        clean
        ;;
    "start")
        start
        ;;
    "stop")
        stop
        ;;
    *)
        echo "unknown command '$1'"
        exit 1
        ;;
esac
