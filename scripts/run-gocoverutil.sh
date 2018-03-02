#!/bin/sh
#
# manually migrate development database to test database and run gocoverutil
#
#  repository: https://github.com/AlekSi/gocoverutil
#
# go get -u github.com/AlekSi/gocoverutil

app=github.com/hyeoncheon/uart

buffalo db drop -e test
buffalo db create -e test
buffalo db schema dump -e development
buffalo db schema load -e test

export GO_ENV="test"
export UART_HOME=`pwd`
gocoverutil -coverprofile=cover.out test -covermode=count $app/... && \
go tool cover -html=cover.out -o cover.html

