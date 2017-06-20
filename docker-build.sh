#!/bin/sh

docker run --rm -v $PWD:/go/src/github.com/epix/dns-bh -w /go/src/github.com/epix/dns-bh golang:1.8 make $1
