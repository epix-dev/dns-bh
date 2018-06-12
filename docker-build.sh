#!/bin/sh

docker run --rm \
    -v $PWD:/go/src/github.com/epix-dev/dns-bh \
    -w /go/src/github.com/epix-dev/dns-bh \
    golang:1.9 sh -c "apt-get update; apt-get install libsqlite3-dev; make $1"
