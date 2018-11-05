#!/bin/sh
#
# manually migrate development database to test database and run gocoverutil
#
#  repository: https://github.com/AlekSi/gocoverutil
#
# go get -u github.com/AlekSi/gocoverutil

app=github.com/hyeoncheon/uart

buffalo pop drop -e test
buffalo pop create -e test
buffalo pop migrate -e test
# bad way while using on test machine:
# buffalo pop schema dump -e development && buffalo pop schema load -e test

GO_ENV="test" \
UART_HOME=`pwd` \
gocoverutil -coverprofile=c.out test -covermode=count ./... && \
go tool cover -html=c.out -o cover.html

