#!/bin/bash

set -e

cfg=$(cat swarm.json)
subnet_name=$(echo $cfg | jq -r ".subnet.name")
subnet_ip=$(echo $cfg | jq -r ".subnet.ip")
num_seeders=$(echo $cfg | jq -r ".numSeeders")
subnet_ip_prefix=$(echo $subnet_ip|cut -d "." -f "1,2,3")"."
tracker_ip=$subnet_ip_prefix"254"

build() {

    subnet_mask=$(echo $cfg | jq -r ".subnet.mask")
    seeder_dbpath=$(echo $cfg | jq -r ".seederDbPath")
    tracker_url="http://"$tracker_ip":23230/"

    ezl_json_fmt="{\"trackerUrl\":\"%s\",\"seederAddr\":\"%s\",\"dbPath\":\"%s\"}"
    ezs_json_fmt="{\"listenAddr\":\"%s\",\"dbPath\":\"%s\"}"

    docker network inspect $subnet_name 1>>/dev/null 2>>/dev/null || docker network create --subnet "$subnet_ip$subnet_mask" $subnet_name

    docker build -t ez_tracker -f dockerfiles/Dockerfile.tracker .

    config_dir=config
    rm -rf $config_dir
    mkdir -p $config_dir
    for i in $(seq 1 $num_seeders)
    do
        ip_suffix=$((i+1))
        printf $ezl_json_fmt $tracker_url $subnet_ip_prefix$ip_suffix":23231" $seeder_dbpath | tpl -t templates/ezl.json.tpl > $config_dir/ezl.json
        printf $ezs_json_fmt ":23231" $seeder_dbpath | tpl -t templates/ezs.json.tpl > $config_dir/ezs.json
        docker build -t ez_seeder_$i -f dockerfiles/Dockerfile.seeder .
    done
    rm -rf $config_dir
}

run() {
    ip=$1
    img=$2
    echo $ip $img
    docker run \
           --rm \
           -d \
           --network $subnet_name \
           --ip $ip \
           --name $img \
           $img
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
