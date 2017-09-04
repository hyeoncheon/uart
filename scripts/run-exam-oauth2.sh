#!/bin/sh

ID="4C0jeMa4mJRWXtjRVVqxD8yoL6c5Rf7rzIC3VWeDIrogjLY3"
KEY="9dqnnT56H7uc1y5w0FPtXG6dnxu5mivnMV3vxrCEQ9MzBes7L3XYSOry83QllPix"
[ "$UART_CLIENT_ID" = "" ] && export UART_CLIENT_ID=$ID
[ "$UART_SECRET_KEY" = "" ] && export UART_SECRET_KEY=$KEY

set -x

cd examples/oauth2

go build client.go
./client
