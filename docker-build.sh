#!/bin/sh

docker run --rm \
    -v $PWD:/go/src/build \
    -w /go/src/build \
    golang:1.14 sh -c "apt-get update; apt-get install libsqlite3-dev; make $1"

