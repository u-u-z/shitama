#!/usr/bin/env bash

export GOPATH=$HOME/go

DIR_SHITAMA=$(pwd)

function build_holder() {
    go build -o ./build/holder/holder ./holder/
}

function build_shard() {
    go build -o ./build/shard/shard ./shard/
}

function build_all() {
    build_holder
    build_shard
    build_client
}

function clean_holder() {
    rm -r ./build/holder/
}

function clean_shard() {
    rm -r ./build/shard/
}

function clean_all() {
    clean_holder
    clean_shard
}
