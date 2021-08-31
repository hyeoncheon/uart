#!/bin/sh
#
# now buffalo support -coverprofile option!

UART_HOME=`pwd` \
	buffalo test -coverprofile=c.out -covermode=count ./... && \
	go tool cover -html c.out -o cover.html
