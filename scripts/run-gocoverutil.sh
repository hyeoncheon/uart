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
buffalo db migrate -e test
# bad way while using on test machine:
# buffalo db schema dump -e development && buffalo db schema load -e test

GO_ENV="test" \
UART_HOME=`pwd` \
gocoverutil -coverprofile=c.out test -covermode=count ./... && \
go tool cover -html=c.out -o cover.html

